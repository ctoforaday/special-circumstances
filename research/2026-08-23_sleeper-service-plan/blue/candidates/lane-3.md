# Lane 3 candidate — local-repo critical-stance lens

**Method lens:** audit the subject artifacts and codebase directly, at the leaf; trust
nothing secondhand — including prior lanes' leaf-reads, which are re-verified here against
the repo at the pinned commit. **Assigned hypothesis order:** hypothesis 3 (cadence /
queueing) first, then breadth.

**Provenance discipline for this lane.** Every plan claim below was re-read from the repo
file at the launch pin, and the pin was verified equal to the working tree so the leaf
reads are trustworthy: `git diff --quiet 2ce929f -- plans/claude-port-plan.md` exits 0
(printed `PLAN MATCHES PIN 2ce929f`). Source note: subject repo `ctoforaday/special-circumstances`,
`plans/claude-port-plan.md` and `plugins/sleeper-service/**` at pin `2ce929f` (PINNED.md,
this run's inputs, accessed 2026-08-23). Two re-runnable checks were written and run;
their scripts live under this candidates dir and are flagged for the synthesizer to anchor
as `proof:`.

---

## Headline (what this lane adds to the union)

1. **The #429 subject is unbuilt at the leaf — mechanically, not by inference.** The
   sleeper-service plugin carries three files (`plugin.json`, `README.md`, `bin/.gitkeep`)
   and none of the four artifacts §3c says it must carry. This **settles the
   verification-status question (lane 2's Q5) directly**: the Phase-4 end-to-end check
   (`claude -p "/self-improve"` → run dir + idea stub) *cannot* have run green, because the
   `/self-improve` command does not exist. No design argument is needed to reach that.
2. **The ship-vs-withdraw question is #429's own title, and the owner has already chosen
   the direction — SHIP** — so this run's job is not to re-litigate the binary but to define
   the leaf-grounded **conditions** under which "ship" is honest, each an implementable tracked
   issue.
3. **Three leaf corrections the "ship" exit must absorb**, or it ships a husk that reads as
   done: (a) the friction input must read the **record/projection**, not `research/*/friction.md`
   — the repo's own record-only design already retired that file; (b) the write-confinement
   guardrail must ship as a **mechanized hook**, not the Phase-4 one-run **observation** it is
   today; (c) cadence for increment 1 is **manual**, and the "daily default" (Resolved decision 6)
   is deferred until (b) is green — manual invocation is exactly what dissolves the queueing
   instability lanes 1–2 found.
4. **Reframing of Q3 from the ungated-store side** (complements lane 2, does not restate it):
   ungated `research/`+`ideas/` writes mean daily cadence builds not an unstable *server queue*
   but a monotonically *growing candidate pool* the human must search to pick a promotion — the
   diverging cost is **selection over the pool** (Q2's ranking problem), and it grows without
   bound absent a decay/archival policy the plan does not specify.

---

## Hypothesis 3 first — cadence and scheduling, at the leaf

**What the repo actually specifies.** Cadence "daily" exists in exactly two prose sites and no
mechanism: Resolved decision 6 ("Self-improve cadence: **daily** default once scheduled (over
the old hourly); manual run always available, scheduling always human-opt-in") and the Phase-4
row's parenthetical ("daily default"). The scheduling *recipes* the plan promises live in
`docs/scheduling.md` (§3c: "default cadence: DAILY; recipes: manual, Windows Task Scheduler +
`claude -p`, cloud agent") — **and that file does not exist** (leaf audit below; `ABSENT
docs/scheduling.md`). So the cadence decision is prose-only; nothing in the repo schedules
anything. Source note: `claude-port-plan.md` §3c / §5 Phase table / Resolved decision 6, pin
`2ce929f`.

**Why the "daily default" is the wrong default for increment 1 — reframed.** Lane 2 modelled
the promotion gate as an M/M/1 server queue and showed ρ≥1 diverges (its `lane2_queue.py`,
Little's Law + Kingman; a strong result this lane does not repeat). This lane audits the
guardrail that the queue model rests on and finds it changes the shape of the risk: the plan's
guardrail is "**the loop writes only `research/` and `ideas/`**" (§3c), and those writes are
**ungated** — they are a corpus, not a blocked queue. So daily cadence does not stall a server;
it **inflates a candidate pool** that the human must search over to choose what to promote. The
cost that diverges is the *selection set*, not the wait time — which routes the cadence risk
straight back into Q2 (there is no defensible ranking, so a growing pool makes an already
unprincipled "pick one" monotonically harder).

Computation (this lane, re-runnable — `blue/candidates/lane3_accumulation.py`; flag for
synthesizer `proof:`). Grounding the per-run yield at the leaf: Phase-4 verify names it in the
singular — "produces a run dir + **idea stub**" — so p ≈ 1 candidate/run, deterministic, not
probabilistic. At daily cadence (7 runs/wk, Resolved decision 6) arrivals are 364 stubs/yr. With
no decay policy specified, the unpromoted pool after one year is `(7 − μ_wk)·52` for a
single-maintainer promotion rate μ: **μ=3/wk → 208 unpromoted; μ=1/wk → 312**. The store already
accumulates — the script measured 28 files / ~448 KB across `ideas/`+`research/` today (this
figure **includes the live run directory**, so treat it as an upper bound on the durable corpus;
`ideas/` alone is 4 files), which already falsifies the plan's "empty scaffolding on publish"
(Resolved decision 4 / §3c repo-layout comment). **Robust, decay-free claim:** pool(t) grows
without bound for any μ<7/wk; **manual/pull cadence sets arrivals = the human's actual promotion
rate, bounding the pool by construction.** Same resolution lane 2 reached (pace by the
bottleneck), reached from the ungated-store side.

**Disconfirming check I ran against my own position.** The plan says daily is *opt-in* with manual
always available (Resolved decision 6). If the human only ever enables the cron when they have
review capacity, arrivals are human-gated and the pool never diverges — the hypothesis weakens to
"documenting daily as the *default* invites leaving an unstable config running." I concede that
weakening: the finding is not "daily is unsafe under all use" but "**daily is the wrong thing to
document as the default for increment 1**, and manual is the right first increment." Notably #429's
own "ship" exit describes a manual `/self-improve` and mentions no cron — so the leaf and the issue
agree.

---

## Breadth 1 — the build-state audit that settles verification status

Re-runnable check (this lane — `blue/candidates/lane3_buildstate.sh`; flag for synthesizer
`proof:`). Output, verbatim:

- actual tree: `plugin.json`, `README.md`, `bin/.gitkeep` — nothing else.
- planned §3c artifacts: `ABSENT skills/continuous-learning/SKILL.md`, `ABSENT
  commands/self-improve.md`, `ABSENT commands/graduate.md`, `ABSENT docs/scheduling.md`.
- `plugin.json` `dependencies` array (§3c: "declared structurally in `plugin.json`'s
  `dependencies` array … not left to `/sc-doctor` to police"): **ABSENT**.
- any suite hook enforcing write-confinement to `research/`+`ideas/`: **NONE**.

This is the same enumeration #429 states in its body ("carries zero code — a README, a
plugin.json, and bin/.gitkeep"), reproduced independently from the filesystem rather than
copied from the issue. Source note: GitHub issue ctoforaday/special-circumstances#429,
https://github.com/ctoforaday/special-circumstances/issues/429 (accessed 2026-08-23).

**Consequence for #429.** The Phase-4 verify column is the operator's prototype-and-test gate for
this plugin, and it names a driveable check. That check is **unrunnable today** — not
"tested-and-inconclusive", not "failed", but **not-testable-because-absent**. validation-loop's
three states must not collapse into "the design looks fine"; the honest state is *the artifact
under test does not exist*, which is a stronger and cleaner thing to say. It means "ship" cannot
mean "merge the working loop" (there is none to merge) — it means **author the first buildable
increment and drive its Phase-4 check to green before the plugin stops being a husk.**

---

## Breadth 2 — the write-confinement guardrail has no mechanism (policy-without-mechanism)

The plan's load-bearing safety promise is "the loop writes only `research/` and `ideas/`;
promotion into rules/skills requires the human (Semantic Consent)" (§3c) — repeated in the
plugin README. At the leaf, **nothing enforces it.** The only hooks the suite ships that touch
this surface are `frank-exchange-of-views/hooks/hooks.json`, whose two binaries do (1) the
blue-`report.md` lockdown and (2) run/identity injection — neither confines the *sleeper loop's*
writes to `research/`+`ideas/`. The Phase-4 verify ("touches only research/+ideas/") is an
**observation of one run**, not a gate; a happy-path observation cannot bound the blast radius of
an autonomous, scheduled, headless loop. This matches red's standing patterns
`policy-without-mechanism` and `happy-path-only gate` (run's staged red-gap inventory,
`inputs/red-gap-patterns.md`, accessed 2026-08-23).

**Why this is a ship-blocker, not a nicety.** The entire argument for letting the loop run
unattended is that its blast radius is mechanically bounded — that is the guardrail that makes
the *other* open design questions (Q1–Q4) tolerable to defer. If the bound is prose, the loop's
one safety guarantee is unverified precisely where it runs without a human watching (auto mode,
headless `claude -p`, the sleeper cron). The fix is cheap and is a natural sibling to the hooks
the suite already ships as tested Go binaries with capability-gating: a `PreToolUse(Write|Edit|
MultiEdit|Bash)` gate that **denies** a loop-context write whose resolved path is outside
`research/`+`ideas/`. Ship it *with* the loop, verified green, exactly as §3a′ argued the qlty gate
must ship with the hook rather than later.

---

## Breadth 3 — the friction-corpus input is doubly refuted at the leaf (corroborates Q1)

Lane 2's Q1 found the Phase-4 row names the loop's input as "the friction records —
`research/*/friction.md`" while friction is really a mediated event. This lane adds the repo's
**own design doctrine** as independent corroboration: `record-only-channel.md` (Decision
2026-07-23) makes the event record the single inter-agent content channel, and its cross-cutting
notes tie in "#87 (friction verb — same 'emit, don't route around' pattern; fold friction-emit
into Stage 1)". So `research/*/friction.md` is not merely under-specified — it is a file the
repo's record-only design **explicitly retired in favour of the friction verb/event**, read back
through a projection. This run is itself the demonstration: there is no `friction.md` in the run
inputs; friction is written with the `friction` verb and read through the projection. Source note:
`plans/record-only-channel.md`, pin `2ce929f`.

A `/self-improve` that "reads the friction records" (as #429's ship exit words it) must therefore
read the **friction event log / its projection**, never `friction.md` prose. Reading prose
reproduces the facts-are-fields silent-zero: once the file stops being written the read returns
zero, indistinguishable from "no friction this cycle", and topic selection starves while the board
looks clean. This is an acceptance-criterion for the ship increment, not a separate finding.

---

## The ship-vs-withdraw resolution for #429 (this lane's synthesis)

**Direction: SHIP — but conditionally, and the conditions are the increment.** #429's title *is*
the ship-vs-withdraw question; its body offers two honest exits (ship the first increment /
withdraw from `marketplace.json` + `PLUGINS` in `scripts/bootstrap-plugins.sh`); and the owner's
own comment (2026-08-23) takes it over as "the 'ship the first increment' exit, made concrete."
Per semantic-consent the ship/withdraw call is the human's, and he has made it — so this lane does
not re-open the binary. It defines what "ship" must contain to be honest, each piece an
implementable tracked issue with an acceptance criterion a skeptic can re-run:

- **Issue A — manual `/self-improve` (no cron).** Enumerate own rules/skills/agents/ideas/research
  → select → `/research` → idea stub in `ideas/`. *Acceptance:* headless `claude -p "/self-improve"`
  produces a run dir + one idea stub. (This is the Phase-4 check, made real for the first time.)
- **Issue B — friction input reads the record, not prose.** `/self-improve` consumes the friction
  **event log / projection** across runs. *Acceptance:* with friction events present the selector
  sees them; with the writer stopped the read **errors or reports "not measured"**, never a silent
  zero (facts-are-fields clause 3).
- **Issue C — write-confinement hook.** A tested, capability-gated `PreToolUse` binary that denies
  a loop-context write resolving outside `research/`+`ideas/`. *Acceptance:* a seeded write to
  `plugins/…` from loop context is **denied**; a write to `ideas/` succeeds; degrades to one stderr
  line when the binary is absent (mirrors the fx bootstrap guard).
- **Issue D — a written selection criterion** (lane 2's Q2): ship a one-line priority rule and
  record chosen topic + score on the record; a bandit/WSJF controller is a deferred enrichment.
- **Issue E — cadence is manual for increment 1; demote "daily default."** Defer the cron /
  `docs/scheduling.md` recipes to a later increment **gated on Issue C green**. *Acceptance:* the
  shipped increment schedules nothing; the plan's Resolved decision 6 is amended to "manual;
  scheduling deferred."
- **Issue F — manifest honesty either way.** Add the `dependencies` array (§3c) so
  sleeper-service structurally requires frank-exchange-of-views; and note the plan is stale on
  plugin count (says three; `marketplace.json` lists four — `gray-area` was added and never
  entered the port plan). *Acceptance:* enabling sleeper-service auto-enables its dep; the port
  plan's plugin census matches the manifest.

**Withdraw (exit 2) is the correct fallback only if A–C cannot be delivered** — but they are cheap
and A is already the plan's own Phase 4. Withdrawing forfeits the one plugin that consumes friction
data no other plugin reads; shipping A–C with the corrections above is strictly better and keeps
the manifest's promise instead of leaving a husk that "reads as done from the manifest" (#429's own
words, the facts-are-fields/complete-the-concept idiom this repo is built on).

---

## Lines of inquiry (this lane)

**Pursued**
- **L3-A build-state leaf audit** (primary breadth). *Hypothesis:* the #429 subject is unbuilt, so
  Phase-4 cannot have run green and "ship" cannot mean merge-the-loop. **Paid off**, mechanically
  (`lane3_buildstate.sh`). Settles lane 2's Q5 from the filesystem rather than by inference.
- **L3-B cadence leaf audit + ungated-pool reframing** (hypothesis 3). *Hypothesis:* daily lives
  only in prose and the instability is better modelled as unbounded candidate-pool growth at the
  human selection gate. **Paid off** (`lane3_accumulation.py`); complements lane 2's server-queue
  model, agrees on the resolution (pace by the bottleneck), and adds the manual-first conclusion.
- **L3-C write-confinement mechanism audit.** *Hypothesis:* the core safety guardrail is prose,
  not a mechanism. **Paid off** — no suite hook enforces confinement; Phase-4 verify is an
  observation. Becomes ship-Issue C.
- **L3-D friction-input doctrine corroboration.** *Hypothesis:* the repo's own record-only design
  already refutes `research/*/friction.md` as the input. **Paid off** (`record-only-channel.md`
  #87). Independent corroboration of Q1; becomes ship-Issue B.
- **L3-E read #429 directly.** *Hypothesis:* the issue text resolves what "ship vs withdraw" means
  and prior lanes could not read it. **Paid off** — read via the GitHub MCP (`issue_read`),
  **closing the GitHub-access gap lane 2 filed as friction**. The issue names both exits and the
  owner's chosen direction.

**Deferred**
- **Decay/archival policy for `ideas/`.** The pool-growth result assumes no decay; whether
  unpromoted stubs should expire/archive is a real design question this lane raises but does not
  resolve (memory-architecture §6 has PRUNE/DECAY for MEMORY, per lane 2's Q4, but nothing covers
  `ideas/`). A later run should size a decay/archival rule against a real promotion-rate sample.

**Weighed and rejected**
- **Re-running lane 2's M/M/1 queue computation.** Rejected as duplicative: lane 2's proof stands
  and this lane is additive-only. The ungated-pool reframing is the non-overlapping contribution.
- **Auditing the four-vs-three plugin drift as a headline.** Rejected as a headline; it is real
  (`marketplace.json` lists `gray-area`, absent from the port plan) but minor and belongs folded
  into ship-Issue F, not elevated over the safety/verification findings.

---

## For the synthesizer

- Two re-runnable checks to anchor as `proof:` against the sentences they back —
  `blue/candidates/lane3_buildstate.sh` (build-state / confinement enumeration) and
  `blue/candidates/lane3_accumulation.py` (candidate-pool growth). Both run from repo root at pin
  `2ce929f`; the pool script's corpus count includes the live run dir (stated as an upper bound).
- One tool-managed citation to add: GitHub issue #429 (title, body's two exits, owner comment),
  https://github.com/ctoforaday/special-circumstances/issues/429, accessed 2026-08-23.
- Subject-artifact references (no URL; cite by pin `2ce929f` per PINNED.md):
  `plans/claude-port-plan.md` §3c/§4/§5/Resolved decisions; `plans/record-only-channel.md`;
  `plugins/sleeper-service/**`; `plugins/frank-exchange-of-views/hooks/hooks.json`.
