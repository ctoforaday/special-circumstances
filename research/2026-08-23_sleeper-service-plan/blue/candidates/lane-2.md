# Blue lane 2 candidate — primary-literature lens

**Topic:** the first shipped increment of the sleeper-service self-improvement loop — inputs, topic
selection, cadence, the promotion ladder and its human gates, prototype-and-test discipline, and the
ship-vs-withdraw resolution for #429.

**Method lens:** primary-literature (papers, specs, standards; leaf sources over commentary). This
lane's job is to bring the established science and standards to bear on the plan's design decisions —
to say which of the plan's choices are corroborated by a body of primary work, which are contradicted
by it, and where the literature is silent so the plan is on its own. Ordering: hypothesis 2
(topic-selection) first, then breadth across Q3–Q1.

Source notes are recorded inline as prose (URL, title, access date). They are NOT citations — the
synthesizer attaches citations with the tool. Access date for every web source below: **2026-08-23**.

A standing caution carried from red's gap-pattern inventory (`inputs/red-gap-patterns.md`): every
figure below is pinned to its exact source AND its experimental condition, because the inventory's
most frequent kill is "right paper, wrong arm" (`within_source_condition_misattribution`) and "real
figure, wrong source" (`citation_status_and_misattribution_patterns`). Where I transfer a result from
the setting it was measured in to the sleeper-service setting, I label it **[inference]** rather than
letting it read as established.

---

## Q2 (primary slice) — Is a defensible topic-selection function required to ship, and does one exist?

### What the plan actually specifies

`self-improve.md` is described as: "enumerate own rules/skills/agents/ideas/research → pick one →
/research 'how should X evolve?' → idea stub w/ alternatives" (claude-port-plan.md §3c, lines
361–362, read at leaf in the run's `plans/`). §4's Phase-4 row adds that it "consumes the friction
records ... as a first-class input." The load-bearing verb is **"pick one"**, and the plan states no
criterion by which one is picked. Consuming an input is not ranking by it: nothing in §3c or §4 maps
the friction corpus, run records, telemetry, or `law/proposed/` rulings to a priority ordering over
candidates. **Confidence: high** that the criterion is unspecified in the plan text — this is a
leaf-read of the two sections the plan itself points at, and a `grep` of the plan for selection terms
(priority/rank/score/select) returns nothing that constitutes a function. (I flag the file-type
blindspot red warns of: the criterion could live in the `continuous-learning` skill body, which does
not yet exist in this pre-implementation plan — so "absent" here means "absent from the design", not
"absent from a shipped artifact". See Lines of Inquiry.)

### Why "pick one" without a criterion is not neutral — the primary literature

The claim I want to make defensible is: **an unguided "pick one" over an enumerated candidate list is
not a neutral draw; it is a biased selector, and the bias is toward salience/recency rather than
value.** Three bodies of primary work bear on this.

1. **Selection under a budget is a solved problem class with known-good strategies, and "pick one
   arbitrarily" is the strategy the field measured against and beat.** Active learning is exactly the
   problem of choosing, from a large pool, the one item whose resolution is most valuable per unit
   cost. The canonical survey enumerates the query strategies that beat random selection —
   uncertainty sampling, expected model change, expected error reduction, variance reduction — and
   the entire field's premise is that *which item you pick* changes the learning rate materially
   versus random. (Source: Burr Settles, "Active Learning Literature Survey", Univ. Wisconsin–Madison
   CS Tech Report 1648, 2009; survey PDF mirror
   https://sfu-db.github.io/cmpt884-fall16/Lectures/884_presentation_on_active_learning.pdf and
   ResearchGate record https://www.researchgate.net/publication/228942691_Active_Learning_Literature_Survey.)
   The transfer to sleeper-service is direct: the loop is an active learner choosing which part of its
   own methodology to interrogate next. **[inference]** that the specific strategies transfer, but
   **[established]** that unguided selection is the baseline these strategies were built to improve on.

2. **When selection is repeated over time under uncertainty, the governing theory is the multi-armed
   bandit, and its central result is that the regret of a good policy grows only logarithmically
   while a policy that ignores the explore/exploit trade-off pays linear regret.** UCB1 achieves
   uniform logarithmic regret over time for any bounded-reward arm distribution. (Source: Auer,
   Cesa-Bianchi, Fischer, "Finite-time Analysis of the Multiarmed Bandit Problem", *Machine
   Learning* 47:235–256, 2002; https://link.springer.com/article/10.1023/A:1013689704352, HAL mirror
   https://inria.hal.science/inria-00574987/document. The logarithmic-regret lower bound is Lai &
   Robbins 1985.) The load-bearing consequence for the plan: a daily loop that "picks one" with no
   explore/exploit bookkeeping has no regret guarantee at all — it can revisit the same few loud arms
   indefinitely (pure exploitation of salience) or wander (pure exploration), and the literature says
   the cost of that is not a constant, it compounds with the number of cycles. **Confidence: high**
   that this is the right formal frame; **[inference]** on the quantitative regret bound applying to
   a topic-selection loop whose reward signal is not yet defined (the plan defines no reward).

3. **The plan's implicit selector — an LLM reading an enumerated list and choosing — has a measured
   positional bias.** "Lost in the Middle" shows that language-model accuracy at *locating and using*
   an item from a list follows a U-shaped curve: items at the start and end of the context are used;
   items in the middle are systematically neglected. (Source: Liu, Lin, Hewitt, Paranjape,
   Bevilacqua, Petroni, Liang, "Lost in the Middle: How Language Models Use Long Contexts", *TACL*
   12:157–173, 2024; https://aclanthology.org/2024.tacl-1.9/, arXiv:2307.03172.) The measured
   condition is retrieval/QA over long contexts, NOT candidate selection — so this is **[inference]**,
   and I mark it as the weakest of the three. But it is a concrete, disconfirmation-resistant reason
   to doubt that "enumerate → pick one" samples the candidate space uniformly: the enumeration order
   itself biases the pick, which means the loop's topic choice is partly an artifact of how the
   candidate list happens to be laid out. That is the opposite of a defensible ordering.

### What a defensible criterion would look like — and that the field already has one for this exact shape

The plan does not need to invent a selection function; the product-development literature already has
the one that fits a backlog-of-improvements-under-limited-review-capacity. **Weighted Shortest Job
First (WSJF) = Cost of Delay ÷ job size** sequences a backlog to maximize economic value delivered per
unit of the scarce resource. (Source: Donald G. Reinertsen, *The Principles of Product Development
Flow*, Celeritas, 2009 — the origin of Cost of Delay as the economic primitive; SAFe's operational
writeup https://framework.scaledagile.com/wsjf/ and the Cost-of-Delay overview
https://en.wikipedia.org/wiki/Cost_of_delay.) Mapped to sleeper-service: Cost of Delay ≈
friction-frequency × blast-radius (how often the gap bites, how much it touches); job size ≈ expected
fix cost. This gives the loop a *stated, auditable* reason for every topic choice — which is precisely
what an unattended, human-gated loop needs, because a choice the loop cannot justify is a choice the
human cannot ratify at the gate. **Confidence: medium-high** that WSJF is the right off-the-shelf
frame; it is a recommendation, not an established fact about sleeper-service.

**Blue's pragmatist caveat (defending against scope creep):** the plan should NOT ship a full bandit
controller in increment 1. The minimum viable criterion is a *written, single-line priority rule*
(e.g. "rank open candidates by friction-frequency × blast-radius ÷ estimated-fix-cost; break ties by
oldest-unaddressed") plus recording the chosen topic and its score on the record so a later run can
audit the ordering. That is a one-issue change with acceptance criteria, and it converts "pick one"
from an unfalsifiable act into a checkable one. The bandit/active-learning machinery is a *deferred*
enrichment, not a ship-blocker — recorded as such in Lines of Inquiry.

### Verdict on Q2

The hypothesis holds in its narrow form: **shipping increment 1 with no written selection criterion is
a defect, because the literature is unanimous that unguided selection is the baseline good strategies
beat, and the plan's implicit selector (LLM-over-a-list) has a measured non-neutral bias.** It does
NOT hold in its strong form ("a bandit controller is required to ship") — that over-scopes. The
resolution is a stated priority rule now, controller later.

---

## Q3 (breadth) — Does daily cadence build an unstable review queue? (computed)

The plan resolved cadence to **daily default** ("over the old hourly", Resolved decision 6, line 438),
with manual always available and scheduling human-opt-in (line 439). The promotion into rules/skills
is human-gated (line 367). That is a producer (autonomous runs) feeding a single-server consumer (the
maintainer at the promotion gate) — a queue, and queues have a stability condition.

**The governing law needs no distributional assumption.** Little's Law (L = λW) relates mean backlog
L, arrival rate λ, and mean wait W for *any* stationary queue — "no probability distribution, no
exponential assumption, no Poisson process required" (Source: Little's Law, derivation notes, Karl
Sigman, Columbia,
http://www.columbia.edu/~ks20/stochastic-I/stochastic-I-LL.pdf; overview
https://en.wikipedia.org/wiki/Little%27s_law). The stability condition is λ < μ (arrival rate below
service rate); at ρ = λ/μ ≥ 1 the queue grows without bound, L → ∞ and W → ∞ (standard M/M/1 result,
same sources plus the M/M/1 utilization treatment).

**I computed ρ across plausible parameters** rather than assert it (script:
`blue/candidates/lane2_queue.py` under the run dir, re-runnable; the synthesizer should anchor this as
a `proof:`). Model: the *gated* resource is promotion into rules/skills — writes to `research/` and
`ideas/` are ungated and accumulate freely (that is a corpus, not a blocked queue). Arrivals λ =
7·p per week, where 7 is the daily cadence and p ∈ [0,1] is the fraction of runs that produce a
*promotable* candidate; service μ = promotions a single maintainer clears per week (bounded by the
hours to read a full multi-round debate report and judge a rule change).

Computed results (see script output, recorded against this run):

- **Stability frontier (daily cadence, ρ<1 requires p < μ/7):** μ=1/wk → stable only if yield
  p < 0.143; μ=2 → p < 0.286; μ=3 → p < 0.429; μ=5 → p < 0.714; μ=7 → p < 1.000. A single maintainer
  clearing ~3 promotions/week is stable **only if fewer than ~43% of daily runs yield a promotable
  candidate.**
- **The blow-up is nonlinear near the frontier (M/M/1, μ=3/wk):** at ρ=0.93 mean backlog ≈ 14 reports
  and mean wait ≈ 5 weeks; at ρ=0.98, ≈ 49 reports and ≈ 17 weeks; at ρ=1.00, ≈ 749 reports and ≈ 250
  weeks. The degradation is a hockey stick, not a ramp — the system looks fine at moderate load and
  then falls off a cliff.
- **Pure daily cadence with every run promotable (p=1, λ=7/wk) is unstable for any single-maintainer
  μ ≤ 7/wk** (ρ = 7/μ ≥ 1).

**What this establishes, stated precisely.** The M/M/1 metrics (L, W) are an *idealization* (Poisson
arrivals, exponential service) — the specific backlog numbers are illustrative, not predictions. But
two things are robust to the distributional assumption: (a) the ρ≥1 divergence is Little's-Law-level,
distribution-free; (b) the near-ρ=1 nonlinearity holds for any queue with variability (Kingman's
formula makes wait grow like ρ/(1−ρ), superlinear near 1, for general G/G/1). So the honest claim is:
**daily cadence as an always-on default is over-provisioned for a single-maintainer promotion gate
unless candidate-yield is deliberately kept low or review is batched.** **Confidence: high** on the
frontier arithmetic (it is p < μ/7, recomputed); **medium** on the real-world μ and p, which are
estimates I did not measure.

**The disconfirming path, honored.** The plan says daily is *opt-in* and manual is always available.
If the human only enables the schedule when they have review capacity, then λ is human-gated by
construction and ρ<1 trivially — the queue never fills. Under that reading the hypothesis weakens from
"daily cadence builds an unstable queue" to "*documenting daily as the default* invites the operator to
leave an unstable configuration running." That weaker claim still stands and still has a fix.

**The resolution the arithmetic points to (and it matches the primary literature on flow):** pace
arrivals by the bottleneck. This is the pull-system / WIP-limit principle from the same Reinertsen flow
body cited in Q2 — cap work-in-progress at the gate and let completion pull the next run, rather than a
calendar pushing runs independent of review capacity. Concretely for increment 1: trigger a run when
unreviewed-corpus depth is *below* a threshold (pull), or gate the daily schedule behind a
"promotion-queue not full" check, instead of an unconditional daily cron. **Confidence: medium-high**
this is the right direction; it is a design recommendation.

---

## Q4 (breadth) — Is the load-bearing human gate the MEMORY write, not the rule-skill promotion?

The plan's stated guardrail: "the loop writes only `research/` and `ideas/`; promotion into
rules/skills requires the human" (line 367). That gate sits on the *rule-skill* rung. The promotion
ladder, however, has a rung *below* it: **insight → MEMORY → rule-skill → cheatsheet** (§3c line 358;
memory-architecture.md §6 makes it physical: trajectory/MEMORY.md note → short-term → active concept →
projections/active.md → rule-skill). The hypothesis: the quiet, compounding failure is the MEMORY
write, because unversioned facts a later run reads as evidence propagate silently.

### The primary literature says self-referential data loops degrade — under a specific condition

**Model collapse is real when generated data REPLACES real data.** Training successive generations on
recursively generated content drives distributional drift (early collapse) and permanent loss of
low-frequency events (late collapse); "within a few generations, original content is replaced by
unrelated nonsense." (Source: Shumailov, Shumaylov, Zhao, Papernot, Anderson, Gal, "AI models collapse
when trained on recursively generated data", *Nature* 631:755–759, 2024;
https://www.nature.com/articles/s41586-024-07566-y, open PDF
https://www.pure.ed.ac.uk/ws/portalfiles/portal/460496122/ShumailovEtalNature2024AIModelsCollapseWhen.pdf.)
The pinned condition matters: the collapse result is for the **replace** regime.

**When generated data ACCUMULATES alongside the real data, collapse is avoided.** The direct rebuttal
paper shows that "accumulating successive generations of synthetic data alongside the original real
data avoids model collapse across a range of model sizes, architectures, and hyperparameters", whereas
replacing does collapse. (Source: Gerstgrasser, Schaeffer, et al., "Is Model Collapse Inevitable?
Breaking the Curse of Recursion by Accumulating Real and Synthetic Data", arXiv:2404.01413, 2024;
https://arxiv.org/abs/2404.01413.) **This is the single most important result for blue's defense of
the plan on Q4**, and it cuts *for* the design: `research/` and `ideas/` are append-only, git-tracked,
never-replacing stores. A loop that accumulates its own outputs beside the original corpus is on the
*safe* side of the collapse literature — provided nothing prunes/replaces the real substrate. So the
naive "self-improvement loops collapse" framing is contradicted; the design's accumulate-don't-replace
shape is corroborated. **Confidence: high** on the accumulate-vs-replace distinction; it is the
explicit finding of both papers read at leaf.

### But MEMORY is not accumulate-only — and that is where the hazard survives

memory-architecture.md's lifecycle (§6) is *not* purely additive: the dream loop **prunes** short-term
notes, **deprecates and deletes** stale concepts, **decays confidence**, and regenerates
`projections/active.md` that later sessions read as always-on context. Deletion + regeneration is a
*replace* operation on the always-on view. That is the regime the collapse literature warns about, and
it is exactly the rung the plan's stated gate does not cover. The failure mode has two named primary-
literature analogues:

1. **Confirmation bias / error accumulation in self-training.** A loop that reads its own prior
   outputs as labels over-fits to its own errors; "confirmation bias is the result of over-fitting to
   incorrect pseudo-labels ... the accumulation of noise arising from utilization of incorrect
   pseudo-labels", and SSL "inherently lacks mechanisms to correct this self-reinforced bias."
   (Source: Arazo, Ortego, Albert, O'Connor, McGuinness, "Pseudo-Labeling and Confirmation Bias in
   Deep Semi-Supervised Learning", arXiv:1908.02983, 2019; https://arxiv.org/abs/1908.02983.) The
   condition is SSL pseudo-labels, so applying it to a MEMORY store read as evidence is **[inference]**
   — but the mechanism (a system reading its own unverified outputs as ground truth, with no external
   correction) is structurally identical to a loop promoting an unversioned MEMORY fact that later
   runs treat as established.

2. **A knowledge store read as evidence is dominated by a tiny number of bad entries.** PoisonedRAG
   shows that injecting **5 malicious texts per target question into a database of millions** yields a
   **90% attack success rate** at making the model emit the attacker's answer. (Source: Zou, Geng,
   Wang, Jia, "PoisonedRAG: Knowledge Corruption Attacks to Retrieval-Augmented Generation of Large
   Language Models", arXiv:2402.07867, USENIX Security 2025;
   https://arxiv.org/abs/2402.07867, https://www.usenix.org/conference/usenixsecurity25/presentation/zou-poisonedrag.)
   The pinned condition is **adversarial injection**, not spontaneous accretion — so I do NOT claim the
   loop poisons itself at 90%. What the result bounds is the *fragility* of the read-as-evidence
   pattern: a handful of wrong entries in a vast store is sufficient to dominate the output. A single
   wrong "prior" written by one bad cycle is, functionally, one injected text — and the plan gates the
   loud rung (a bad RULE) while leaving this quiet one open. **[inference]** on the transfer;
   **[established]** on the underlying fragility figure in its adversarial condition.

### The repo already ruled on this — and the ruling is a primary source

`priors-are-poison.md` (ruling 2026-07-19, read at leaf) independently reached the same place from
inside the system: seeding a chair with cross-run scorecard "memory" is "not neutral, it is likely
HARMFUL" — three named mechanisms (Goodhart, cross-topic confound, salience) and "no evidence of
benefit." Its distinction is the load-bearing one for Q4: **in-run feedback on the current question is
legitimate; cross-run priors from other questions are poison.** The constitutions enforce the same —
"memory is a checklist, not a library", "every fact re-verifies at the leaf." So the plan's own
doctrine already says a MEMORY store read back as evidence across runs is the hazard; the port plan's
guardrail simply does not place a gate there. **Confidence: high** — this is a leaf read of the repo's
own ruling.

### Verdict on Q4

Hypothesis holds in a refined form: **the accumulate-only `research/`/`ideas/` writes are safe (and
positively corroborated by Gerstgrasser 2024), but the MEMORY rung — because its lifecycle prunes,
deletes, decays, and regenerates an always-on projection later runs read as evidence — is a
replace-regime, self-referential channel that the plan's stated gate does not cover.** The fix is not
necessarily a human gate on every MEMORY write (that would defeat the point of unattended
consolidation); it is a *write-discipline* that keeps MEMORY on the safe side of the collapse
literature: never let generated/derived facts replace the real substrate; keep provenance; require
corroboration (≥2 independent trajectories, which memory-architecture.md §6 already proposes) before a
fact becomes always-on; and re-verify at the leaf on read (which the constitutions already require).
Blue's pragmatist read: much of this discipline is *already in* memory-architecture.md — the gap is
that the *port plan's guardrail sentence* under-describes it, gating only the rule rung. The concrete
increment-1 issue is: state the MEMORY-rung write-discipline in the guardrail, not just the rule-rung
human gate.

---

## Q5 (breadth) — #429 resolves on verification STATUS, not design merit

The operator brief makes prototyping-and-testing a *binding* process requirement: "blue's process MUST
verify its work and recommendations by prototyping and testing ... Recommendations the run could have
cheaply falsified and didn't are audit findings" (OPERATOR-BRIEF.md, Standing process requirement).
The repo's own discipline (validation-loop skill) is "only the command's output closes it" and names
three distinct states — not-tested, tested-and-inconclusive, tested-green — that a reader acts on
differently.

**The primary-literature frame is the verification/validation distinction, and it is standardized.**
IEEE Std 1012 defines verification as confirmation by objective evidence that specified requirements
are fulfilled (the product is built right), and validation as confirmation that requirements for the
intended use are fulfilled (the right product is built). (Source: IEEE Std 1012, "IEEE Standard for
System, Software, and Hardware Verification and Validation", 2016 ed.,
https://ieeexplore.ieee.org/document/8055462; ANSI overview
https://blog.ansi.org/ansi/ieee-1012-2016-verification-validation-vv/; 1012-1998 text mirror
https://people.eecs.ku.edu/~hossein/Teaching/Stds/1012.pdf.) The Phase-4 verify column names a
concrete verification check: headless `claude -p "/self-improve"` produces a run dir + idea stub and
touches ONLY `research/`+`ideas/` (claude-port-plan.md §Part 5, Phase 4 row, line 410). That is a
verification obligation with objective, re-runnable evidence — exactly what the standard requires the
"build it right" claim to rest on.

**Self-improvement is licensed by a verifier, per the primary literature — which is why Q5 is the
keystone.** STaR shows a model *can* bootstrap its own capability by generating candidate rationales
and **keeping only those that reach the correct answer** — the correctness filter is what makes the
loop converge instead of drift. (Source: Zelikman, Wu, Mu, Goodman, "STaR: Bootstrapping Reasoning with
Reasoning", NeurIPS 2022, arXiv:2203.14465; https://arxiv.org/abs/2203.14465.) Generalized: an
autonomous self-improvement loop is safe to run unattended to the exact extent that its outputs pass
through a verifier the loop cannot game. For sleeper-service the verifier IS the human promotion gate
plus the prototype-and-test discipline plus the write-confinement invariant. So the ship-vs-withdraw
question for #429 is not "is the design elegant" — it is "has the Phase-4 verification actually run
green, and is write-confinement mechanically enforced?"

**What I could not verify — and it is filed as friction.** I could not read #429's actual contents
(no GitHub access from this seat) and found no evidence in the run inputs that the Phase-4 end-to-end
check has been *run and observed green* — Phase 4 in the plan carries no ✅ (Phases 0 and 1 do; 2–5 do
not, line 406–411). Under the standard and the repo's own validation-loop, an unrun verification means
the "build it right" claim is in the **not-tested** state, which is different from failed and different
from passed. **The honest resolution, stated as a decision rule rather than a verdict I cannot
license:**

- If the Phase-4 E2E has NOT been run green and write-confinement is unproven → the increment is in
  the not-tested state; **withdraw-or-hold** is the honest call regardless of how sound Q1–Q4 make the
  design look (collapsing not-tested into "design looks fine" is the precise move validation-loop
  forbids).
- If the Phase-4 E2E runs green AND write-confinement (touches only `research/`+`ideas/`) is
  demonstrated → **ship is licensed** even with Q1–Q4's design questions open, because the guardrail
  that bounds blast radius is mechanically enforced and the remaining questions are improvements, not
  safety-blockers.

**Confidence: high** on the decision rule (it follows directly from the standard + validation-loop);
**this lane cannot resolve which branch obtains** because reading #429 and running the E2E are both
outside its reach. That is the accurate state to hand the synthesizer: the resolution is *conditional
and computable*, and the run should discharge it by actually driving the E2E, not by arguing design
merit.

---

## Q1 (breadth) — Input-channel integrity: are the loop's inputs read from records or from prose?

This hypothesis is primarily a doctrine-vs-implementation audit rather than a primary-literature
question, so this lane treats it lightly and defers the deep census to the doctrine-lens lane. The
primary-literature contribution is the general principle behind facts-are-fields: **schema-on-write
(validate at the write, reject bad values) vs schema-on-read (recover a fact from text shape at read
time)** is a long-standing data-engineering distinction, and the failure mode facts-are-fields names —
a no-match returning a plausible zero indistinguishable from an honest zero — is the "silent failure"
class that observability literature treats as the most dangerous because it is invisible to the
consumer.

The load-bearing leaf fact: the plan's Phase-4 row names the loop's input as "the friction records —
`research/*/friction.md`" (line 410). But in *this very run*, friction is not a markdown file — it is
the `friction` verb writing an event on the record, read through a projection (confirmed by leaf: the
record tool's help exposes a `friction` verb and no `friction.md` exists in the run;
`inputs/` contains `red-gap-patterns.md` and scorecards, not a `friction.md`). Reading the friction
CORPUS back out of `research/*/friction.md` prose is the "document standing in for a record" failure
facts-are-fields describes: once that file stops being written, the read returns zero, which reads
identically to "no friction this cycle", and topic selection silently starves. **This directly
compounds Q2:** a selection function that ranks by friction-frequency is only as trustworthy as the
channel it counts friction on; if the count is recovered from prose, a broken pipe looks like a clean
board. **Confidence: high** that the plan text names a prose file while the engine uses a mediated
record; **the disconfirming leaf** (does the plan intend `friction.md` to be a GENERATED projection of
the friction record, staleness-gated?) is the check that would collapse the hypothesis — and the plan
does not say either way, so the honest state is "under-specified, and the safe reading is
record-mediated inputs only." I defer the full input-by-input census (run records, telemetry,
`law/proposed/` rulings) to the doctrine lane; the primary-literature lane's contribution is only the
schema-on-write framing and the Q1↔Q2 compounding.

---

## Lines of inquiry — pursued, deferred, rejected

**Pursued (carried into the draft above):**

- **Q2 via active-learning + bandit + WSJF primary literature** — pursued and paid off: the literature
  is unanimous that unguided selection is the beatable baseline, giving the "no written criterion is a
  defect" claim a real evidential spine rather than an assertion. Narrowed the claim from "needs a
  bandit controller" (over-scoped) to "needs a written priority rule now, controller deferred."
- **Q3 via queueing theory + a computation** — pursued and paid off: Little's Law gives a
  distribution-free stability condition, and the computed frontier (p < μ/7) plus the near-ρ=1 hockey
  stick quantify exactly when daily cadence breaks. The disconfirming reading (opt-in ⇒ human-gated
  arrivals) survived and reshaped the claim rather than killing it.
- **Q4 via model-collapse + RAG-poisoning + self-training-bias literature** — pursued; the surprise
  was that the strongest result (Gerstgrasser accumulate-vs-replace) cuts FOR the plan's append-only
  corpus, so the honest draft corroborates the design on the `research/`/`ideas/` rung and localizes
  the hazard to the MEMORY rung's replace-regime lifecycle.
- **Q5 via IEEE 1012 + STaR** — pursued; produced a conditional decision rule for #429 keyed on
  verification status, which is the operator's binding prototype-and-test requirement made formal.

**Deferred (worth a later run, not this lane):**

- **A full bandit/active-learning controller spec for topic selection** — deferred: increment 1 should
  ship the one-line WSJF-style rule; the controller is a later increment once a reward signal exists.
  A later run should pick this up once the loop has run enough cycles to define reward empirically.
- **Kingman's-formula / G/G/1 modeling of the review queue with measured variability** — deferred: the
  distribution-free divergence and the M/M/1 illustration are enough to make the design point; a
  variability-aware model wants real μ/p measurements the run does not have.
- **The full input-by-input facts-are-fields census (Q1) across run records, telemetry, and
  `law/proposed/` rulings** — deferred to the doctrine lens lane, which owns that audit; flagged the
  Q1↔Q2 compounding so it is not lost.

**Weighed and rejected:**

- **Framing Q4 as "self-improvement loops inevitably collapse" (the Shumailov-only reading)** —
  rejected: it is contradicted at leaf by Gerstgrasser 2024's accumulate-vs-replace result, and
  asserting it would have been a "real figure, wrong condition" error of exactly the kind red's
  inventory kills. The accurate frame is regime-dependent (replace collapses; accumulate does not).
- **Applying PoisonedRAG's 90% figure as a spontaneous self-poisoning rate** — rejected: that figure
  is an *adversarial-injection* attack-success rate, not an accretion rate; using it as the latter is
  `within_source_condition_misattribution`. Kept only as a bound on the *fragility* of read-as-evidence
  stores, in its true condition.
- **Treating #429 as a design-merit question answerable from the plan alone** — rejected: it collapses
  validation-loop's three states into one and pre-empts a check the run can actually drive. Left as a
  conditional decision rule instead.

---

## Confidence summary (calibration, for the synthesizer)

| Claim | State | Confidence |
|---|---|---|
| Plan specifies no topic-selection criterion (leaf-read §3c/§4) | established (in the *design*, pre-implementation) | high |
| Unguided "pick one" is the beatable baseline (active learning / bandits) | established in source domains; **[inference]** to this loop | high / medium |
| LLM-over-a-list has positional selection bias | **[inference]** from Lost-in-the-Middle (measured for retrieval, not selection) | medium |
| WSJF is the right off-the-shelf criterion | recommendation | medium-high |
| Daily cadence ⇒ ρ<1 only if yield p < μ/7; hockey stick near ρ=1 | computed (script recorded) | high (arithmetic) / medium (real μ,p) |
| Accumulate-not-replace corpus is on the safe side of model collapse | established (Gerstgrasser 2024, leaf) | high |
| MEMORY rung's prune/decay/regenerate lifecycle is a replace-regime hazard the plan's gate misses | derived from memory-architecture.md §6 + collapse literature | medium-high |
| PoisonedRAG bounds fragility of read-as-evidence stores (adversarial condition) | established in its condition; **[inference]** to accretion | high / medium |
| #429 resolves on verification status via a conditional decision rule | derived from IEEE 1012 + validation-loop | high (rule) / unresolved (which branch) |
| Q1 plan names prose `friction.md` while engine uses a mediated record | established (leaf) | high |

## Friction (this lane)

- **No GitHub access from this seat** — could not read issue #429's actual contents to confirm the
  ship-vs-withdraw framing against the real bug text; worked around it by treating #429 as a
  verification-status decision rule keyed on the Phase-4 E2E, but the leaf (bug text, current
  labels/state) is unread. What I would have done with `gh`/GitHub read: fetch #429, confirm it is an
  open bug (red's `citation_status` pattern: "open bug" that is actually closed-not-planned), and pin
  the exact acceptance criteria the report's issues must satisfy.
- **Registration hook not injecting run dir** (engine-wide, per register output) — recorded once.
- **`prove` cannot anchor at lane stage** — the queueing computation is real and recorded under the
  run dir (`blue/candidates/lane2_queue.py`), but `prove` requires a `--quote` matching `blue/report.md`,
  which is still a stub; the proof can only be anchored at synthesis. Not a blocker, noted so the
  synthesizer anchors it rather than re-deriving.
