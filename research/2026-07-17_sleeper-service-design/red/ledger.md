# red/ledger.md — sleeper-service design run (SINGLE SOURCE OF TRUTH for gap status)

Round 2 merge, 2026-07-17. Lens passes: L1–L4 (leaf citation verification, 4 slices),
L5 (logic & completeness), L6 (dark-side & risk). Candidate files preserved under
`red/candidates/round-2-lens-*.md`; found_by is auditable against them.
Grading legend: likelihood × impact × complexity-to-mitigate → severity.
"likelihood: certain" on textual defects = the defect is verified present (the mass
mapping's convention); consequence-likelihood is carried in impact.

Round-2 headline: all 30 round-1 repairs verified as new claims (repair-regression
discipline). 29 gaps CLOSED (16 clean, 13 with regression — successors below carry
declared lineage). The round-2 board is dominated by defects IN the round-1 remediation
machinery: the canary, the origin-tag, the snapshot, the plugin copy, the resume cap.

## OPEN GAPS (23)

### R1-7 — inputs/PINNED.md pins a path that does not exist at the pin (infrastructure — standing with the LEAD)
- found_by: [L2] (blue self-disclosed in [^PortPlan]; also standing friction)
- location: `inputs/PINNED.md` row "`plans/claude-port-plan.md` | `7bc501e`" vs
  [^PortPlan] — "the pin's `plans/claude-port-plan.md` path does not exist ... at `7bc501e`".
- problem: `git cat-file -e` → MISSING at pin. Setup tooling asserted a nonexistent path;
  port-plan citations are snapshot-grade, not pin-grade. Not a blue defect; blue's
  disclosure is accurate and was re-verified round 2 (L1: working-tree quotes verbatim at
  `6df52af`, clean). Awaiting lead ruling/fix; carried unchanged.
- required_fix: fix setup-script pin validation or stage the port plan into inputs/.
- grading: certain × low × low → severity **low**

### R2-1 — the denial canary's actor, observer, and abort mechanism are unspecified; as printed the abort is an instructional gate of the class §4.1 refutes (supersedes: R1-28)
- found_by: [L5]
- supersedes: [R1-28]
- location: §2.2 step 0 — "First in-session action: DENIAL CANARY — attempt one
  out-of-fence write, abort the run unless DENIED." vs the same step's header "PREFLIGHT
  (wrapper, OUTSIDE the model session ...)"; §4.3 layer 2 — "fence ENFORCEMENT is proven
  per run by the wrapper's step-0 denial canary".
- problem: the write attempt must come from the model session (that is what is tested) but
  the abort is attributed to the wrapper, and no mechanism connects them. `claude -p`
  takes one prompt; the JSON envelope (with `permission_denials`) arrives at session END —
  a post-hoc check has already let the whole run execute. If the abort is the model's own
  responsibility, it is a prompt-level MUST — exactly the gate class §4.1 refutes with the
  report's own evidence ("the MUST-try clause has no observable"; 0/175). Workable
  mechanisms exist (stream-json two-phase drive: wrapper sends canary prompt, parses the
  deny event, only then sends the real prompt; or a probe-P2-style micro-session with an
  argued same-environment guarantee) — none is specified, and the two anchors contradict
  each other on who acts.
- required_fix: specify the two-phase drive (or probe-session variant + same-env argument)
  and reconcile the step-0 header with the "in-session" label.
- grading: medium-high × medium-high × low → severity **medium-high**

### R2-2 — the canary cannot isolate the hook fence: layers 1 and 2 fence the SAME boundary, so a fence-dormant run with live permissions passes the canary — the exact drift R1-28 was minted to catch (supersedes: R1-28)
- found_by: [L5, L6] (independent; identical mechanism analysis)
- supersedes: [R1-28]
- location: §4.3 layer 2 — "so marker loss (scheduler misconfig, shell nesting, rung
  migration ...) fails CLOSED instead of silently dropping to the permission-rules-only
  configuration H4 refuted (R1-28)"; §8 OQ2 — "the smoke test must also verify the canary
  itself aborts on a deliberately fence-dormant run".
- problem: every out-of-fence canary target is also outside layer 1's allow set
  (`Edit(<REPO>/research/**)`, `Edit(<REPO>/ideas/**)` under `dontAsk` auto-deny + explicit
  denies), so whenever layer 1 is live the canary is DENIED regardless of hook state. On
  marker loss (hook dormant, operator settings intact — the modal drift case) the canary
  passes and the run proceeds under layer 1 alone — H4's refuted configuration — now with
  a false "enforcement PROVEN per run" stamp (§0 tree). Because the two layers fence the
  identical boundary, no probe write exists that layer 1 allows and layer 2 denies: the
  deny-outcome canary architecturally cannot isolate the hook layer. OQ2's own acceptance
  test ("canary aborts on a deliberately fence-dormant run") FAILS as designed — the test
  and the claim contradict; one is wrong. Verification-probe-layer-masking class.
- required_fix: a POSITIVE per-run hook-liveness signal independent of deny outcome (hook
  writes a "fired + denied path X at ts" record the wrapper confirms non-empty; or wrapper
  parses deny PROVENANCE — PreToolUse-hook block vs permission-rule denial — from
  stream-json, if distinguishable: verify). If neither is buildable, downgrade §0/§4.3
  claims from "fails CLOSED / proven per run" to "at least one deny layer is live" and
  re-grade the marker-loss residual; fix OQ2's test to match whichever claim survives.
- grading: medium × medium-high × low-medium → severity **medium-high**

### R2-3 — step 4's FEOV invocation is never reconciled with "removes ALL script execution from the session": one branch stalls the nightly driver, the other falsifies risk row 4's re-argued acceptance (supersedes: R1-16, R1-21)
- found_by: [L5]
- supersedes: [R1-16, R1-21]
- location: §2.2 step 4 — "invoke the frank-exchange-of-views machinery at smoke scale
  (1 lane, 1 round, 1 citation pass, bulk model ...)"; vs §4.3 layer 4 — "the round-1
  shape (R1-16) removes ALL script execution from the session"; vs §6 row 4 — "the
  session's Bash surface is now pinned-argv read-only git commands — no script execution,
  no model-chosen subprocess write targets".
- problem: FEOV is scripted by doctrine ([^ResearchCommand]: "the mechanics are
  scripted"; capture emits cost.md + run-record-audit.md). The report never states HOW
  step 4 executes under the sleeper profile. Branch (a): FEOV scripts run via session
  Bash → auto-denied under the git-only allowlist → the loop's core research step stalls
  every night. Branch (b): workflow scripts execute through a harness-side runner outside
  the Bash permission surface → "removes ALL script execution" and row 4's
  channel-narrowness leg ("the only subprocess reachable is the pinned git set") are both
  false — the model reaches a script-running workflow surface nightly, and R1-21's
  re-argued acceptance (leg a) rests on a premise step 4 negates. One branch is true by
  construction; whichever it is, a round-1 conclusion fails.
  Sibling-repair-composition class: R1-16's hardening and step 4's requirement were each
  verified alone, never together.
- required_fix: determine the execution locus by probe; then either enumerate FEOV's
  script needs wrapper-side/pinned-argv, or re-scope row 4 + layer 4 to cover the
  workflow-runner channel (argue hook coverage over its writes) and re-argue the
  acceptance on the true premise.
- grading: high × high × low-medium → severity **high**

### R2-4 — §2.2 step 2 still executes a script IN-SESSION; R1-16 propagation miss (supersedes: R1-16)
- found_by: [L5]
- supersedes: [R1-16]
- location: §2.2 step 2 — "SCORE (script; cheap tier only if scoring needs language): the
  §1.4 formula; log the full scored table to the run dir so the pick is auditable."
- problem: under the §4.2 profile a session-side script invocation at step 2 is
  auto-denied and the loop stalls — the failure mode R1-16(d) named. The repair exists in
  the report's own text: §1.4 stage 1's wrapper-staged docket already carries a `score`
  column, so step 2 is either wrapper-work done at step 0 (say "read the staged scored
  table") or model-side language-tier scoring (drop the word "script").
- required_fix: reword the step to match the wrapper-side architecture.
- grading: medium × low-medium × trivial → severity **low-medium**

### R2-5 — the authorship cap is keyed on a dir-NAME glob; the loop's own step-4 FEOV sub-run lands friction in a differently-named dir that is untagged AND supplies the "non-sleeper corroboration" the cap requires (supersedes: R1-25)
- found_by: [L6]
- supersedes: [R1-25]
- location: §1.5 — "harvest.mjs tags every row sourced from a `*_self-improve/` run dir
  `origin: sleeper` ... and requires at least one non-sleeper occurrence before a
  sleeper-origin class enters the top-3 comparison"; §2.3 research field — "<run-dir of
  the bounded pass; verdict stamp>".
- problem: the nightly run creates TWO run dirs: `research/<date>_self-improve/` and the
  spawned bounded FEOV pass's own run dir (topic-slug named per FEOV convention; the
  report never says it inherits the `_self-improve` suffix). Harvest parses "every run's
  friction.md" — both. Only the first matches the glob. The bounded pass's loop-authored
  friction is untagged, counts at full recurrence × staleness, and SATISFIES the
  "at least one non-sleeper occurrence" gate — the loop's own spawned sub-run provides the
  "independent corroboration" the cap was built to demand from outside the loop. The
  mechanism is circumvented by the loop's default nightly control flow, not an edge case;
  §6 row 10's re-grade ("premise made true by mechanism") rests on the mechanism holding.
  Origin-tag-naming-keyed class.
- required_fix: tag by CAUSAL PROVENANCE, not name — the wrapper stamps every run dir it
  (or its session) spawns with an `origin: sleeper` marker file at creation; harvest reads
  the marker, not the dir name. Restate §1.5; re-grade row 10 after the sub-run
  corpus-membership is closed.
- grading: medium × medium × low → severity **medium-high** (load-bearing guard
  circumvented by normal operation; cheap to fix — same convention as R1-25's grade)

### R2-6 — "requires non-sleeper corroboration — or, failing that, flags it" is an undecided disjunction; the two branches are materially different mechanisms and each horn carries an unowned cost (supersedes: R1-25)
- found_by: [L5, L6] (L5: the undecided-"or"; L6: the hard-bar horn's suppression cost)
- supersedes: [R1-25]
- location: §1.5 — "requires at least one non-sleeper occurrence before a sleeper-origin
  class enters the top-3 comparison — or, failing that, flags it `sleeper-only` in the
  stub provenance so the human sees the loop citing itself"; §6 row 10 — "requires
  non-sleeper corroboration (or a visible `sleeper-only` provenance flag) before a
  sleeper-origin class reaches the pick".
- problem: "requires X — or, failing that, flags" is a label, not a requirement. Hard-bar
  reading: a class observable ONLY in sleeper runs (headless `-p`, `--bare`,
  MCP-under-headless, the scheduled profile — the loop's own operating path, which
  human-present runs do not exercise) can NEVER reach the pick — the loop is structurally
  blind to its own most-relevant failures, and the suppression widens as the corpus
  becomes sleeper-dominated. Flagged-pass reading: the loop CAN pick its own
  uncorroborated complaint and containment reduces to the human noticing a provenance flag
  at stub-triage — under §1.1's own Dependabot base rate, a weak reed. Row 10's
  Low-Medium re-grade silently assumes the strong reading while the text licenses the
  weak one. False-equivalence-disjuncts class.
- required_fix: decide the branch; state the chosen horn's cost (e.g. severity-gated
  bypass: an aborted run / fail-closed preflight / hook crash class enters flagged
  `sleeper-only`; ordinary classes need corroboration); make row 10's grade track the
  decision.
- grading: high × medium × trivial → severity **medium**

### R2-7 — the wrapper start/end snapshot has no defined reader or comparison step — the undefined-receiver defect R1-29 fixed for loop death, reintroduced for tamper evidence (supersedes: R1-26)
- found_by: [L5, L6] (independent; concurring fix shape)
- supersedes: [R1-26] (also destabilizes former row-4 leg (b) — carried into R2-3's
  re-argument scope)
- location: §4.3 layer 5 — "a write-then-revert inside the run shows as a start/end delta
  the model cannot reach, restoring the durable tamper-evidence property the DGM analogy
  actually requires"; §6 row 4 revisit trigger — "any start/end snapshot delta"; §3.4
  dead-man — "`/prosthetic-conscience:doctor` reports 'last successful sleeper run: N days
  ago'".
- problem: shows — to whom, when? No wrapper auto-compare, no abort/flag on delta, no
  doctor line, no review cadence. Run N+1 fires before any human looks; a delta can sit
  unread indefinitely. Row 4's acceptance leg (b) and its revisit trigger presuppose a
  reader the design does not specify. Detection-that-nobody-reads is not detection —
  window-without-a-watchman; the DGM analogy's own force is that the lineage was INSPECTED.
- required_fix: wrapper compares this run's start/end (and end-vs-prior-start) guardrail
  hashes at step 7; on mismatch raise a doctor-visible flag ("sleeper guardrail hash
  changed between runs N and N+1") and/or fail the next preflight closed. Until then,
  row 4's leg (b) and revisit trigger are unbacked.
- grading: low-medium × medium-high × low → severity **medium**

### R2-8 — the operator-owned plugin copy has no creation/refresh owner or freshness check; "the human-gated plugin update path" names a path defined nowhere; preflight asserts LOAD, not VERSION (supersedes: R1-15)
- found_by: [L5, L6] (independent; L6 adds the merged-fix-runs-undeployed consequence)
- supersedes: [R1-15]
- location: §3.2 — "`--plugin-dir` pointing at an operator-owned, read-only plugin copy
  outside the write surface (e.g. `~/.claude/sleeper/plugins/`, refreshed only by the
  human-gated plugin update path — NEVER the repo's `plugins/` working tree"; §3.2 —
  "Step 0 asserts sleeper-service + frank-exchange-of-views are present and aborts loudly
  if absent."
- problem: R1-15's pin mints a THIRD copy (distinct from the repo and the
  `~/.claude/plugins/` cache) and leaves its lifecycle unbuilt: no named creator, no
  concrete refresh recipe (the marketplace update dance covers the cache, not this bespoke
  dir), no staleness detection. Consequence: a guardrail fix merged to main via layer 5's
  PR gate does NOT reach the executing copy until a human remembers an undefined step —
  the fence the nightly run executes can lag every hardening indefinitely, believed
  deployed. Preflight asserts plugins LOAD, not that they are the intended version;
  nothing compares the copy's hash to an expected value. Self-referential-repo-drift
  class: a merge landing ≠ the executing surface fixed.
- required_fix: preflight pins and verifies the copy's content hash (or committed version
  stamp) against the operator-approved value, fail-closed on mismatch; scheduling.md gains
  the refresh step in the guardrail-PR merge checklist; doctor gains a "sleeper plugin
  copy: hash mismatch / N commits behind" line.
- grading: medium × medium × low-medium → severity **medium**

### R2-9 — the dead-man surface is pull-only: "a surface the human already looks at" is asserted, not evidenced, and the loop's adoption motive removes exactly that babysitting (supersedes: R1-29)
- found_by: [L6]
- supersedes: [R1-29]
- location: §3.4 dead-man — "the surface the human already looks at, so 'missed day
  (harmless)' and 'loop dead three weeks' are distinguishable without inventing a new
  monitoring channel"; §6 row 15 — "doctor dead-man line ... (a surface the human already
  reads)".
- problem: doctor is human-invoked; the dead-man line is read only when the operator
  chooses to run it, and the point of an unattended loop is to stop routine checking. If
  doctor is not run for three weeks, "loop dead three weeks" is exactly as invisible as
  with no line at all. R1-29 rejected undefined "loudly" and replaced it with a reader
  whose reading cadence is undefined and counter-incentivized.
- required_fix: a push surface on an N-day-silent threshold (wrapper emits through a
  channel the operator passively receives), or state honestly that death-detection latency
  equals the operator's doctor cadence and grade row 15's residual on that.
- grading: medium × low-medium × low → severity **low-medium**

### R2-10 — resume cap k=3 bounds waste per run-DIR, not per-CAUSE: a deterministic root cause re-wedges every fresh dir nightly until the monthly cap trips, every month (supersedes: R1-29)
- found_by: [L6]
- supersedes: [R1-29]
- location: §3.4 idempotence — "the next fire mints a fresh dir, so a wedged run (corrupt
  dir, dead session id, deterministic abort) cannot be resumed nightly forever, burning
  budget with zero output until the monthly cap trips."
- problem: a root cause that survives a fresh dir (corrupt harvest input, persistent
  qmd/MCP failure, poisoned staged input, deterministic abort in the picked topic) wedges
  the new dir the same way: up to k×(per-run budget) ≈ $15/night with zero output until
  the ~$50 cap trips (~3 nights), resetting monthly. The "cannot ... forever" claim holds
  per-dir only; the sentence's own escape clause concedes the waste.
- required_fix: per-CAUSE dead signature (hash of abort reason); after M consecutive
  fresh-dir deaths with the same signature, HALT the loop and raise the dead-man flag;
  soften the claim to its per-dir scope.
- grading: low-medium × low-medium × low → severity **low-medium**

### R2-11 — 30-day auto-stale composes with human graduation latency to churn good-but-queued stubs: the same class is re-researched every 30 days (supersedes: R1-22)
- found_by: [L6]
- supersedes: [R1-22]
- location: §1.4 stage 2 — "an older untriaged stub auto-stales — harvest ... re-enters
  the class in the docket flagged `stub-stale`"; §2.4 — "$150–400 class spend ... once per
  graduation, not once per day."
- problem: a GOOD stub queued for graduation (a heavy human-present event; weeks of
  latency is normal) auto-stales at 30 days, re-enters the docket, gets re-picked and
  re-researched, minting a fresh stub whose age resets — indefinitely. The window is tuned
  against gate-inattention noise but is shorter than realistic graduation lead time, so it
  penalizes the good case.
- required_fix: distinguish "untriaged" from "triaged: graduation-queued" (a human-set
  `status: queued` exempts from auto-stale while still counting for dedupe); or make the
  window configurable against observed graduation cadence.
- grading: medium × low × low → severity **low-medium**

### R2-12 — the idempotency/resume gate still reads loop-writable state; R1-19's explicitly-named sibling got neither the fix nor the risk-argument (supersedes: R1-19)
- found_by: [L5]
- supersedes: [R1-19]
- location: §2.2 step 0 — "today's run-dir idempotency check (exists+complete → exit 0;
  exists+incomplete → resume, capped at k=3 attempts, then mark dead + fresh dir next
  fire)"; §3.4 — "the wrapper writes a `DEAD` marker plus a dated abort record".
- problem: R1-19's own text named it ("Same shape, lower stakes: the idempotency gate
  reads loop-writable `research/**` state"); the fix relocated the ledger and stopped. The
  completeness signal the exit-0/resume/fresh decision reads lives in the run dir
  (loop-writable), and the DEAD marker's location is unstated (row 15 puts abort RECORDS
  operator-owned, but the marker the next fire's check consumes is naturally in-dir). A
  run that mangles its own completeness state (a partial write at abort — no malice
  needed) flips the wrapper's decision on model-authored evidence. Stakes low (bounded by
  k=3 + per-run budget) — likely the right risk-accept, but the report's own discipline is
  that acceptance is argued, not silent.
- required_fix: wrapper derives completeness from its own step-7 record in the
  operator-owned dir; state the DEAD marker's location; or argue the risk-accept
  explicitly.
- grading: low-medium × low × low → severity **low**

### R2-13 — §6 row 5's likelihood cell still carries the pre-R1-9 flat "(no API)" that §5.1 now contradicts as over-broad (supersedes: R1-9)
- found_by: [L3]
- supersedes: [R1-9]
- location: §6 risk matrix row 5 — "No programmatic quota pre-check | **Certain (no
  API)** | Low ... | RISK-ACCEPT with §5.1's layered static guards".
- problem: R1-9's requalification (spend limits: no API; RATE limits: API-readable,
  Admin-key-only) propagated to §5.1, the H5 verdict, [^ConsoleLimits], and new
  [^RateLimitsAPI] — but not to this cell. The conclusion (nothing a subscription-auth
  scheduler can poll) is unchanged and correct; the parenthetical justification is stale.
  Incomplete-repair class.
- required_fix: "(no spend-limit API; rate-limit API unreachable at this auth tier —
  §5.1/R1-9)".
- grading: certain × trivial × trivial → severity **low**

### R2-14 — §7 self-audit still characterizes pricing as MEDIUM after the R1-11 upgrade to leaf-verified HIGH; the §7 upgrade list omits R1-11 (supersedes: R1-11)
- found_by: [L4]
- supersedes: [R1-11]
- location: §7 Pattern B/E bullet — "pricing figures graded MEDIUM with canonical source
  named; no number laundered."
- problem: §5.2 and [^Pricing] both now assert leaf-verified HIGH (R1-11); §7's
  Round-1-update paragraph enumerates the banked upgrades (STOP, [^UsageAPI],
  [^AIScientist], issue statuses) and omits pricing. A reader auditing confidence from §7
  alone carries a stale MEDIUM. Incomplete-repair / self-audit-lag class.
- required_fix: append "(upgraded to leaf-verified HIGH round 1, R1-11)" and add R1-11 to
  §7's upgrade list.
- grading: certain × trivial × trivial → severity **low**

### R2-15 — default-rung drift: §1.4 makes rung 0 "the DEFAULT and may be terminal"; §3.4's ladder still stamps rung 1 "RECOMMENDED default" unqualified (supersedes: R1-14)
- found_by: [L5]
- supersedes: [R1-14]
- location: §3.4 ladder rung 1 — "**RECOMMENDED default** — the only local option where
  every flag is explicit and version-pinnable"; vs §1.4 — "**rung 0 — manual
  `/self-improve`, same code path, zero standing cost — is the DEFAULT and may be
  terminal**".
- problem: the rung-0 row does say "the default until the human opts into scheduling", so
  the residue is one unqualified label two rows down — but §3.4's table is what an
  implementer reads, and the two sections answer "what ships as default?" differently as
  printed. Body-lags-repaired-section class.
- required_fix: "RECOMMENDED default AMONG SCHEDULED RUNGS, once the human opts in".
- grading: certain × trivial × trivial → severity **low**

### R2-16 — gate-survival table, rung-0 column: compound L2 cell claims YES for a canary that cannot exist without the wrapper; rung-0 (default-mode) spend never enters the monthly ledger (supersedes: R1-27)
- found_by: [L5]
- supersedes: [R1-27]
- location: §3.4 gate-survival table — row "L2 hook fence + step-0 denial canary", R0
  cell "YES (cache copy)"; row "Wrapper controls ...", R0 cell "n/a (human present)".
- problem: (a) the L2 row is compound (fence + canary); at rung 0 there is no wrapper,
  hence no canary — the same table's wrapper row says so; honest cell: "fence YES (cache
  copy); canary n/a". The table exists (R1-27) precisely to stop rung cells overstating
  gate presence. (b) the ledger is wrapper-written; rung-0 manual runs bypass it, so in
  the design's DEFAULT mode run costs are never appended — later opt-in starts against a
  ledger blind to prior manual spend; mixed months undercount. Fair posture for rung 0
  alone, but the composition with the cap arithmetic is unstated.
  Exhaustive-sweep-omits-own-hard-case class.
- required_fix: split the R0 L2 cell; one sentence declaring manual-run spend
  out-of-ledger by design (or wrapper-wrap the manual path).
- grading: certain × low × trivial → severity **low**

### R2-17 — §3.3(2a) grants and revokes the pdf/arxiv tool surface in one sentence: under `--strict-mcp-config` naming qmd only, ToolSearch has no pdf/arxiv servers to discover — and the preloaded research-protocol skill's MUST-try becomes structurally unsatisfiable nightly
- found_by: [L5]
- location: §3.3 item 2 — "(a) the loop's MCP profile is `--strict-mcp-config
  --mcp-config <sleeper-mcp.json>` naming **qmd only** (fewer servers is the #68375
  mitigation; research subagents reach pdf/arxiv tools via ToolSearch per the shipped
  seat grants)".
- problem: ToolSearch discovers deferred tools from DECLARED servers; the flag the
  parenthetical annotates removes the declaration (strict-mcp-config ignores the project
  `.mcp.json` that declares pdf-reader/arxiv-latex). Step 4 preloads research-protocol,
  whose clause makes any PDF-hitting citation pass "an incomplete audit" without trying
  those tools — a MUST silently violated every night. Either the sleeper MCP config names
  the pdf/arxiv servers too (re-opening the #68375 fewer-servers trade the same sentence
  makes) or nightly stubs carry a standing degraded-citation caveat — an acceptable trade,
  but it must be chosen, stated, and reflected in the stub contract's confidence field.
- required_fix: pick the trade; one sentence in §3.3 + one stub-contract note.
- grading: certain × low-medium × low → severity **low-medium**

### R2-18 — $2–5/night ceiling × 30 nights = $60–150/month against a ~$50/month cap: normal operation EXPECTS the monthly guard to trip mid-month, and the dead-man surface cannot distinguish the resulting skip streak from death
- found_by: [L5]
- location: §5.2 tier table — "whole daily run — `--max-budget-usd` **$2–5**; monthly
  ledger cap ~$50 (operator-tunable)"; §1.4 — "$2–5/night ceiling (ledger-capped ~$50/mo)".
- problem: at the stated ceiling the cap binds between day 10 ($5) and day 25 ($2) —
  the guard trips routinely unless typical spend is well under $1.67/night, and typical
  spend is never stated (the smoke shape "prices in single-digit dollars" is compatible
  with both sides). Composition: §6 row 6 treats skips as harmless anomalies and row 15
  distinguishes "missed day" from "loop dead three weeks" — a cap tripping at day ~20
  manufactures a legitimate multi-day skip streak monthly, which the doctor line cannot
  tell from death without surfacing the skip REASON. Unreconciled-numeric-floors class
  (recomputed, not restated: 5×10=50; 2×25=50).
- required_fix: state expected per-run spend from the measured smoke figure; size the cap
  to cadence × expected (with headroom) or declare cap-trip the intended month-end
  throttle AND teach the dead-man line the skip reason.
- grading: medium-high × low-medium × trivial → severity **low-medium**

### R2-19 — §0's artifact enumeration omits the new skill file and plugin manifest its own tree prints six lines above
- found_by: [L1]
- location: §0 — "exactly THREE new code artifacts ... plus two command prompts and a
  scheduling doc; everything else reuses shipped machinery".
- problem: the §0 tree introduces `skills/continuous-learning/SKILL.md` and
  `.claude-plugin/plugin.json` — new artifacts that are none of code artifact, command
  prompt, or scheduling doc; "everything else reuses shipped machinery" is literally false
  for those two entries. Exhaustive-sweep-omits-own-specimen class.
- required_fix: make the enumeration total over the printed tree ("+ a skill file and the
  plugin manifest").
- grading: certain × low × trivial → severity **low**

### R2-20 — [^Backlog] pin range "15–17" under-covers the assemble-on-failure sub-claim by one line
- found_by: [L1]
- location: Footnotes, [^Backlog] — "15–17 (smoke mode ~50k tokens; assemble-on-failure)".
- problem: at the pin, line 18 = (c) ASSEMBLE-ON-FAILURE; the cited range stops at 17.
  Line-scheme validated by two independent anchors (27c = line 27; item 39 = line 39 of a
  39-line file). Both sub-claims' content verified verbatim — pin-navigability defect only.
- required_fix: range → "15–18".
- grading: certain × trivial × trivial → severity **low**

### R2-21 — "The DGM analogy is exact" overstates: DGM's archive admission is not validation-GATED, and the sentence invites reading DGM as precedent for threshold-gated admission
- found_by: [L2]
- location: §2.4 — "The DGM analogy is exact and is the design argument: DGM only admits
  a change to its archive after empirical validation against a benchmark, never on the
  proposer's say-so.[^DGM]"
- problem: evaluation does precede archive entry (abstract: "empirically validates each
  change"), but per the round-1 leaf read (abs+html, citation ledger line 74) admission =
  compile + retained edit-ability; benchmark performance does not gate admission and low
  scorers are deliberately retained (the open-ended-exploration point). Blue's /graduate
  is pass-required — STRICTER than DGM on exactly the dimension the sentence leans on.
  Directionally fine; "exact" is false. Within-source-condition class.
- required_fix: drop "exact" ("direct"); one clause: "DGM evaluates every change before
  archiving but admits even low scorers for exploration — our promotion gate is stricter:
  pass-required." No H2 re-grade needed.
- grading: certain × low × trivial → severity **low**

### R2-22 — §5.2 attaches one [^CostRecord] marker to two figures from different artifacts
- found_by: [L3]
- location: §5.2 — "run 3 was $149.95 [minority: lane-3/local-probe — the run-3
  figure].[^CostRecord]".
- problem: $414.97 is from cost.md; $149.95 is from plans/efficiency-phase.md §I — the
  footnote itself discloses the split, and both figures are verified HIGH.
  Footnote-over-attribution class, honest variant; no reader who follows the footnote is
  misled.
- required_fix: add [^EfficiencyPlan] beside the run-3 figure at next touch.
- grading: certain × trivial × trivial → severity **trivial**

## CLOSURE INDEX

R1-1 | closed | backlog count corrected to 25/39, recount verified at pin (L1 HIGH) | —
R1-2 | closed | scope-fusion split into two attributed sources, both verified at pin lines 27c/31h (L1 HIGH) | —
R1-3 | closed | Sakana quote moved to [^DGMSakana]; ICLR-2026 tag dropped (L1 verified) | —
R1-4 | closed | [^SICA] venue now as the page states; re-fetched verbatim r2 (L1 HIGH) | —
R1-5 | closed | #32191 restated CLOSED-duplicate in §3.3/§6/footnote; propagation grep clean (L2, merge grep) | —
R1-6 | closed | exit-code claim softened to any-nonzero-is-failure; cli-reference re-fetched r2, still no exit table (L2) | —
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
R1-21 | closed_with_regression | row 4 re-argued without actor-benignity as demanded; leg (a) premise now contested by the step-4 locus gap | R2-3
R1-22 | closed_with_regression | 30-day auto-stale mechanism specified in §1.4/§2.3; composes badly with graduation latency | R2-11
R1-23 | closed | Batch demoted to FUTURE note; ≤24h sub-claim since resolved HIGH on the batch-processing page (L3 V2) | —
R1-24 | closed | OQ3 qualifier carried inline in layer-6 table row (merge direct read) | —
R1-25 | closed_with_regression | §1.5 covers authorship, origin-tag + cap + corroboration gate specified; tag is name-keyed and the disjunction undecided | R2-5, R2-6
R1-26 | closed_with_regression | wrapper start/end snapshots to operator-owned log specified; no reader/comparison defined | R2-7
R1-27 | closed_with_regression | per-rung gate-survival table added, rung-3/4 adoption graduation-grade; rung-0 cells overstate | R2-16
R1-28 | closed_with_regression | step-0 denial canary added, marker-loss framed fail-closed; canary mechanism unspecified AND cannot isolate the fence layer | R2-1, R2-2
R1-29 | closed_with_regression | resume cap k=3 + DEAD marker + doctor dead-man line added; pull-only reader + per-dir-not-per-cause bound | R2-9, R2-10
R1-30 | closed | mirror line count corrected to 1,557; merge grep: only §7 grep-log token mention of 1,558 stands | —

## NOTES — upgrades blue may bank (not gaps)

- **[^WindowsHang] MEDIUM → HIGH** (L2 r2): #66395 body fetched; v2.1.169 changelog quote
  + exact regression span (2.1.161 2026-06-02 → 2.1.168 2026-06-06); the footnote's
  "body not fetched" caveat is retirable.
- **[^WebSandbox] surveyed → HIGH** (L2 r2): sandbox-environments page verbatim on
  isolated Anthropic-managed VM + token-outside-sandbox proxy (doc names the GitHub token
  specifically).
- **[^MissedRun] anacron MEDIUM → HIGH** (L2 r2): man7.org anacron(8) verbatim.
- **[^Pricing] ≤24h Batch sub-claim MEDIUM → HIGH** (L3 r2): batch-processing page —
  "Batches expire if processing does not complete within 24 hours." Pricing page
  re-fetched live r2: zero drift on every figure incl. the +30% tokenizer boundary.
- **`--json-schema` ≥2.1.205 and rung-3 ~973MB** pinned HIGH at their leaves (L2 r2).
- **OQ7 raw material** (L2 r2, sandbox-environments page): the built-in sandboxed Bash
  tool "does not support native Windows. On Windows hosts, use WSL2 or one of the
  container or VM approaches"; `@anthropic-ai/sandbox-runtime` (whole-process wrap) is
  beta — §4.3(b)'s posture is the documented one. Same page: "Auto mode replaces the
  prompt with a classifier" — raw material for OQ17's `disableAutoMode` leaf-verify.
- **STOP "~page-of-code"** (L1 r2): MEDIUM color paraphrase; ar5iv body fetched, literal
  phrase absent, Figure 2 (seed improver) is a ~page listing — direction corroborated, no
  gap.
- **Probe P1/P2 residue**: disposition of record unchanged (re-run-and-commit at build);
  not triable from an audit seat.
- **Cross-cutting watch (for the LEAD, recommend-not-block):** L6 observes the wrapper now
  hosts five gate-side controls (preflight, ledger, canary, snapshot, resume cap) and is
  itself entirely trusted — R2-1/R2-2/R2-7/R2-8 are instances of one shape: a wrapper
  control that records or attempts but never verifies-and-surfaces its own outcome. If
  round 3 patches wrapper controls one at a time again, consider a stated invariant:
  "every wrapper gate emits a checked, human-surfaced liveness/outcome record"
  (missing-root-invariant watch; severity declining, so recommend-not-block).
