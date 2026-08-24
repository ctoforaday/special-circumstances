# Lane 1 — adversarial-disconfirming-first

**Method lens.** I hunted evidence AGAINST each frontier hypothesis before evidence for it,
working primarily at the leaf: the repo files at the pinned commit `2ce929f`, the live
frank-exchange-of-views engine's own behaviour during this run, and red's accumulated
gap-pattern inventory as a free checklist. Each finding below is tagged by state —
**CONFIRMED-at-leaf**, **PARTIALLY-DISCONFIRMED**, or **DISCONFIRMED** — and I keep
validation-loop's three verification states distinct: *not-tested*, *tested-inconclusive*,
*contradicted* never collapse into "unclear".

**Pin discipline.** Every repo claim is verified at the pinned evidence base `2ce929f`
(`git ls-tree`/`git show` at that commit), not at the worktree HEAD `9bacb3b` — the two differ,
which red's [self-referential repo drift] pattern warns about. For `plans/claude-port-plan.md`
and `plans/memory-architecture.md` the two commits are byte-identical (`git diff --stat 2ce929f
HEAD` empty), so the drift does not touch this lane's evidence; stated so a later reader need not
re-check.

---

## Q1 — Input-channel integrity (facts-are-fields against the loop's inputs)

**Hypothesis under test:** topic-selection quality is bounded by input-channel integrity; if the
increment reads the friction corpus and judicial rulings from prose (`research/*/friction.md`)
instead of from mediated event records, the read returns a plausible zero when the recording pipe
breaks — the loop looks healthiest exactly when it has gone blind.

**Disconfirming leaf test (Q1's own):** is `friction.md` a GENERATED projection of the friction
record (legitimate — then the hypothesis collapses) or a hand-written file (the facts-are-fields
failure)? The leaf answers **neither**. No `friction.md` data file exists anywhere in the engine:
a whole-tree `find … -name friction.md` returns only the CLI *help page*
`plugins/frank-exchange-of-views/tools/internal/cli/seat/help/friction.md` — documentation, not a
data channel. Friction is a **verb/event**: `record-only-channel.md` (#62, 2026-07-23) retired
every parallel agent→agent content channel and made the event record the sole channel; #87 folded
friction-emit into it; the research-protocol skill states friction "is recorded on the RECORD via
the friction verb" and "the read is the OPERATOR's, on the operator's own surface." This very run
produced **no** `friction.md` and I recorded friction through the verb.

**Verdict:** the hypothesis is **DISCONFIRMED as a description of the current engine** and
**CONFIRMED as a latent defect in the PLAN TEXT.** `claude-port-plan.md` §4 (Phase 4) still
literally names the first-class input as *"the friction records — `research/*/friction.md`, agents'
envelope-reported capability gaps"*. That line is a **stale carrier speaking the pre-#62 model**
([doctrine vs implementation]): implementing it literally re-introduces exactly the "document
standing in for a record" failure the repo already ruled against — reading `research/*/friction.md`
(a file the engine no longer produces) returns a plausible zero indistinguishable from "no friction
this cycle", and topic selection silently starves. This is a **complete-the-concept half-state**:
record-only-channel landed the concept; the Phase-4 row was never updated.

**Corroboration the corrected model is already intended:** THIS run's operator brief (2026-08-23)
redefines the inputs as *"the friction corpus — `friction`/`friction-none` events across runs,
envelope harvests; run records; telemetry; `law/proposed/` rulings"* — record-mediated, not
file-mediated. The fourth named input (judicial rulings) is likewise a record: `law/README.md`
shows rulings arrive through the `motion` verb into `law/proposed/` as PERSUASIVE and are promoted
to AFFIRMED by a human — so the same rule applies (read rulings from the motion record /
`law/proposed/`, never scraped from prose).

**Red-checklist pre-empt:** [verification file-type blindspot] — the "no friction.md" claim rests
on a whole-tree `find`, not a `*.md` grep, so it also covers the compiled/tool layer.

**Tracked-issue shape (FR):** rewrite `claude-port-plan.md` §4 and the `/self-improve` spec to read
the friction corpus from the `friction`/`friction-none` EVENT log and its operator projection, run
records, telemetry, and `law/proposed/` rulings via the `motion` record — naming NO `*.md` file as
an input channel. Acceptance: (a) the spec references no `research/*/friction.md`; (b) a
zero-input case surfaces as a LOUD "not measured / no friction recorded this window", never a silent
zero folded into the healthy case.

---

## Q2 — The missing topic-selection function

**Hypothesis under test:** the scarce resource is a defensible ranking, not candidate supply; the
increment may ship only if it specifies an explicit topic-selection function; absent a written
criterion, autonomous "pick one" defaults to recency/salience and produces churn plus starvation.

**Leaf:** the only specification of selection is §3c — *"enumerate own rules/skills/agents/ideas/
research → pick one → /research"* — and the README's *"pick one"*. No criterion is stated in either.

**Disconfirming path (Q2's own), and it is CLOSED at the leaf:** the criterion might live in the
`continuous-learning` skill. That skill is **UNBUILT** — `git ls-tree 2ce929f | grep
continuous-learning` is empty and the skill directory does not exist. So no selection criterion
exists anywhere in plan or code. Q2 is **CONFIRMED**.

**Argument to the other seats:** in a mature repo candidate supply is effectively unbounded, so the
binding scarce resource is a *defensible ordering* over candidates. "Pick one" with no criterion
defaults an LLM to recency/salience — it re-selects whatever the current context most vividly
surfaces (systematically the last thing touched), producing churn (the same loud topics revisited)
and starvation (quiet high-impact debt never selected). A loop that cannot justify WHY it chose a
topic cannot be left to run unattended; this is a top ship-blocker for the autonomous increment.

**Pragmatist caveat (blue's duty — defend against scope creep):** the fix is cheap, not a research
programme. The operator brief already supplies the raw material — friction *recurrence* across
runs, run records, telemetry, `law/proposed/` rulings. A minimal defensible ranking is
recurrence × blast-radius × (1/fix-cost), tie-broken by most-cited open gap across runs. That is
the difference between shippable and not, not gold-plating.

**Interaction with Q1 (see NEW-A below):** raw friction *frequency* is the wrong count — this run
shows one infra gap (the hook) recurring; a naive most-frequent selector would lock onto it every
cycle. The score must count **distinct causes**, not raw events ([identity-keyed detector /
per-instance cap]).

**Tracked-issue shape (FR):** `/self-improve` MUST compute and log a ranked candidate list with a
stated scoring function before "pick one". Acceptance: a dry run prints the ranked list plus each
score component; the selection is the argmax or a logged human override, so the choice is auditable.

---

## Q3 — Cadence as a queue (computed)

**Hypothesis under test:** the binding constraint is human review throughput at the single
mandatory promotion gate; at realistic single-maintainer service rates a DAILY calendar cadence
drives utilization ρ > 1, accumulating unreviewed research until the gate becomes nominal — so
correct cadence is pull/threshold-triggered, not calendar-based.

**Computation (recorded, rerunnable — `scratchpad/rho.py`; ρ = λ/μ, arrivals = candidate-bearing
runs/week, service = careful human promotions/week). Every input is a stated assumption, not a
measurement; the point is the threshold structure.**

| scenario | λ/wk | μ/wk | ρ | stable? |
|---|---|---|---|---|
| self-improve daily, human promotes 1/wk | 7 | 1 | 7.00 | UNSTABLE |
| self-improve daily, human promotes 2/wk | 7 | 2 | 3.50 | UNSTABLE |
| self-improve daily, human promotes 3/wk | 7 | 3 | 2.33 | UNSTABLE |
| self-improve + dream, both daily, 2/wk | 14 | 2 | 7.00 | UNSTABLE |
| self-improve + dream, both daily, 5/wk | 14 | 5 | 2.80 | UNSTABLE |
| pull/threshold cadence, promotes 3/wk | 3 | 3 | 1.00 | critical |
| weekly cadence, promotes 3/wk | 1 | 3 | 0.33 | STABLE |

At ρ > 1 the unreviewed pile grows at (λ−μ)/week: daily self-improve at 2 promotions/week is
50 reports deep in ~10 weeks; both daily loops at 2/week, ~4 weeks. Only pull/threshold or weekly
cadence brings ρ ≤ 1.

**Disconfirming counter (Q3's own), and it partly lands:** scheduling is *"always human-opt-in,
manual always available"* (Resolved decision 6); daily is a DEFAULT, not a forced arrival rate. So
the instability is **CONDITIONAL** on the human enabling the daily schedule. Q3 weakens from "the
loop is unstable" to **"the DEFAULT cadence is mis-set for a human-gated loop, and the instability
is real whenever the default is accepted."** Verdict: **PARTIALLY-DISCONFIRMED** in force,
CONFIRMED in direction.

**Two compounding leaf facts that strengthen the residual finding:**
1. Unreviewed research accumulates in `research/`, which the repo `CLAUDE.md` marks **gitignored** —
   so the backlog is invisible to PR review and lost on container recycle (only `run-archive/`
   survives). A queue you cannot see is worse than one you can ([gitignored ≠ absent]).
2. `memory-architecture.md` §7.6 ships a **second** daily loop (`/dream`) feeding the *same* single
   human gate — arrivals to the promotion gate are the SUM of both loops (the λ=14 rows).

**Tracked-issue shape (FR):** change the DEFAULT cadence from calendar-daily to **pull/threshold-
triggered** (fire when unreviewed-corpus depth crosses N, or when the human opens the gate); keep
opt-in daily and manual as ceilings. Acceptance: the scheduling doc's default recipe is
threshold/event-triggered, and a computed ρ ≤ 1 at the documented default service rate.

---

## Q4 — The load-bearing gate is the MEMORY write, not the rule-skill promotion

**Hypothesis under test:** because unversioned facts written to a MEMORY store that later runs read
as evidence propagate silently across every future cycle (prior-poisoning), a plan that gates only
promotion-into-rules has gated the visible rung and left the compounding one open.

**Leaf — and note two DIFFERENT ladders are named, which is itself an unresolved finding:**
- `claude-port-plan.md` §3c ladder: *insight → MEMORY → rule-skill → cheatsheet*; guardrail
  *"the loop writes only `research/` and `ideas/`; promotion into rules/skills requires the human."*
- `memory-architecture.md` (a **Proposal** that explicitly *supersedes* the port plan's ad-hoc
  memory position) ladder: *trajectory/MEMORY.md → `short-term/` → active concept (`knowledge/`) →
  `projections/active.md` → CLAUDE.md → rule-skill.* The plan never says which governs the first
  increment.

**Disconfirming path (Q4's own):** MEMORY might be design-rationale-log only, never read back as
evidence → prior-poisoning impossible. **PARTIALLY TRUE** for the repo's own `MEMORY.md`
(design-rationale, PII-scrubbed) but **FALSE for the memory-architecture design**: promotion
`short-term → active concept` is **AUTOMATED** by the dream loop on `review_count ≥ 2` — where
"corroborated" means "seen in ≥ 2 trajectories", a bar that two repetitions of the *same* error
clear — and active concepts are injected into **every session** via a `CLAUDE.md` `@`-import (§5).
Only the FINAL rung (→ rule-skill) is human-gated (§6).

**Verdict: CONFIRMED and sharpened.** The plan gates the LOUD rung (a bad rule shipping) and leaves
the QUIET compounding rung — a wrong "active concept" propagating into every future session's
context — **automated and ungated**. The repo's own `priors-are-poison.md` (2026-07-19) is the
ruling that cross-context priors injected into a chair are harmful (Goodhart, cross-topic confound,
salience — *"the gap-patterns were read and the warned patterns committed anyway — priming the
failure, not preventing it"*). That is the same mechanism one level up: a GLOBAL "active concept"
learned on topic A is injected into topic B's context — exactly the cross-topic confound
priors-are-poison names. memory-architecture mitigates via project/global scoping, but the GLOBAL
`active.md` is injected regardless of topic, so the confound survives for global concepts.

**Two further leaf facts:**
- The guardrail *"writes only research/ and ideas/"* is a **[policy without mechanism]**: it names
  no enforcer. `sc-secrets-gate` covers *outbound secrets*, not *write-path confinement*; no hook
  confines the loop's writes. Prose, not a gate.
- **[authorship evades never-edit guard]:** the loop AUTHORS `research/` + `ideas/` (and, under
  memory-architecture, `short-term/` notes) that LATER cycles harvest as input. "Never edits its
  inputs" is satisfied while amplification runs through *authorship* — the loop feeds on its own
  output. This is the sharpest structural risk in the increment and neither plan addresses it.

**Tracked-issue shapes (FR):**
(a) Put a write-discipline gate on the MEMORY rung: promotion `short-term → active` (and any write
that lands in always-on context) is human-gated OR requires corroboration from ≥ 2 **independent**
trajectories (distinct sessions/topics, not the same error twice), with provenance recorded.
Acceptance: a seeded duplicate-error pair does NOT auto-promote.
(b) Make "writes only research/+ideas/" MECHANICAL — a PreToolUse write-path confinement hook.
Acceptance: a seeded write to `plugins/` from a loop agent is blocked, not merely discouraged.
(c) Resolve which memory ladder (§3c vs memory-architecture) is authoritative for the first
increment; they are different and both shipped as named designs.

---

## Q5 — #429 resolves on verification status, not design merit (the decisive line)

**Hypothesis under test:** prototype-and-test discipline licenses SHIP only on an observed-green
Phase-4 end-to-end run; absent that evidence the honest resolution is WITHDRAW regardless of how
sound Q1–Q4 make the design look.

**Leaf — the entire increment is UNBUILT at the pinned commit:** the sleeper-service README
self-declares *"Status: scaffold only (Phase 0)"*; `git ls-tree 2ce929f -- plugins/sleeper-service`
returns exactly three files (`.claude-plugin/plugin.json`, `README.md`, `bin/.gitkeep`); no
`self-improve`, `continuous-learning`, `graduate`, or `scheduling` artifact exists anywhere in the
tree. Phase 4's own verify check — *"Headless `claude -p "/self-improve"` produces a run dir + idea
stub; touches only research/+ideas/"* — **cannot have been observed green because `/self-improve`
does not exist.** This is unambiguously **not-tested** (the artifact under test is absent), a
different and stronger statement than "tested-inconclusive" or "the design looks unsound".

**Live-verified rung-portability finding (strengthens the not-ship case):** the scheduling story is
headless `claude -p "/self-improve"` via Task Scheduler / cron / cloud. This run demonstrated,
**live**, that the engine's PreToolUse hook did NOT fire — `register` returned *"THE ENGINE'S HOOK
IS NOT REACHING THIS RUN … EVERY seat in it is affected"*, forcing a manual `--seat-id` on every
call. `agent-guardrails` states the suite "runs where prompts don't fire — auto mode, headless,
scheduled sleeper-service loops" and must hold there. A loop whose guardrails ride hooks that are
observably rung-fragile is not safe to run unattended ([gate-stack rung non-portability]; and
[per-instance cap not per-cause] — a deterministic hook miss recurs on every daily instance and no
cross-run detector keyed on stable ids will fire, [identity-keyed detector lineage-blind]).

**Resolution recommendation for #429** (conditioned — see gap below):
1. **Preferred: re-scope #429 to "ship the PLAN + the tracked FRs," not the loop.** What this run
   can stand behind is the *corrected design* (Q1–Q4 fixes) expressed as tracked issues with
   acceptance criteria — which is exactly what the operator brief asks the deliverable to be. The
   loop itself is WITHDRAWN-as-increment because there is nothing built to ship and nothing tested.
2. **If #429 must remain a build issue:** it stays OPEN/blocked with the Phase-4 E2E named as the
   gating acceptance test, PLUS a per-rung gate-survival table (does each guardrail fire under Task
   Scheduler / cron / cloud / manual?), because the live hook failure proves the rung matters.

Either way, SHIP-the-loop-now is not a defensible disposition on the evidence.

**Capability gap (filed as friction):** I could not read #429 itself — no `gh`/GitHub access this
seat. The resolution above is conditioned on #429 being the sleeper-service build bug the operator
brief says this run takes over (FR `ctoforaday/special-circumstances#534` "takes over bug #429").

---

## Lines of inquiry considered beyond the frontier (breadth)

- **NEW-A — the friction channel as a POISONED selection input.** Since the loop's top input is the
  friction corpus and this run shows the friction hook failing, a daily cadence of headless runs
  each emitting hook-absence friction would flood the corpus with one recurring infra complaint;
  a naive "most-frequent friction" selector would select IT every cycle (churn lock). *Hypothesis
  if it pays off:* topic-selection needs de-duplication/decay on the friction corpus — a recurring
  cause counted once, not once per run. *Fate:* PURSUED; folds into Q2's scoring-function
  recommendation (score distinct causes, not raw events).
- **NEW-B — telemetry as a selection input.** The operator brief lists telemetry among inputs;
  telemetry is a per-round trend projection. Considered whether selection should prefer topics where
  prior runs' quality was declining. *Fate:* DEFERRED — plausible but the increment has no
  telemetry-across-runs store yet; a later run should pick it up.
- **CONSIDERED-REJECTED — full M/M/1 waiting-time model** (Wq, not just ρ). Rejected as
  over-precision: the arrival process is calendar-timed, not Poisson, and the finding (ρ>1 ⇒
  unbounded backlog) needs only the utilization ratio. Recorded so a later run does not mistake the
  missing queueing rigor for an unexplored avenue.

## What I could NOT establish (accuracy over completeness)

- **#429's actual contents** (no GitHub access) — the resolution is conditional, not asserted.
- **Whether the operator has ever enabled the daily schedule** — would decide whether Q3's
  instability is live or latent; not in the pinned corpus. *not-tested.*
- **Which memory ladder (§3c vs memory-architecture) is authoritative** for the first increment —
  both are named; the plan does not say. *contradicted-by-omission, not resolved.*

## Sources (prose notes for the synthesizer to cite once with the tool)

- `plans/claude-port-plan.md` §3c, §4, Phase 4 table, Resolved decision 6 — repo pinned `2ce929f` —
  leaf read 2026-08-23.
- `plans/memory-architecture.md` §5 (native-surface projection), §6 (promotion ladder, automated
  `review_count ≥ 2`), §7.6 (daily dream cadence) — pinned `2ce929f` — 2026-08-23.
- `plans/priors-are-poison.md` (ruling 2026-07-19: cross-run priors poison; Goodhart / cross-topic
  confound / salience) — pinned `2ce929f` — 2026-08-23.
- `plans/record-only-channel.md` (#62, 2026-07-23: the record is the only inter-agent channel;
  friction-emit folded in, #87) — pinned `2ce929f` — 2026-08-23.
- `research/2026-08-23_sleeper-service-plan/inputs/OPERATOR-BRIEF.md` (this run's scope:
  record-mediated inputs) and `inputs/run-config.json` — 2026-08-23.
- `research/.../inputs/law/README.md` + `law/precedents.md` (judicial-rulings input: `motion` →
  `law/proposed/` PERSUASIVE → AFFIRMED human gate) — 2026-08-23.
- `plugins/sleeper-service/README.md` ("Status: scaffold only (Phase 0)") and
  `git ls-tree 2ce929f -- plugins/sleeper-service` (three scaffold files) — 2026-08-23.
- `research/.../inputs/red-gap-patterns.md` — checklist lines cited inline by bracketed name —
  2026-08-23.
- Computation `scratchpad/rho.py` (ρ = λ/μ utilization table) — recorded, rerunnable — 2026-08-23.
- LIVE observation: `feov-record … register` output *"THE ENGINE'S HOOK IS NOT REACHING THIS RUN"* —
  this run, 2026-08-23.
