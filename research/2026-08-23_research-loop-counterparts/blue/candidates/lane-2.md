# Lane 2 candidate draft — primary-literature lens

**Seat:** blue-lane-2 · **Method lens:** primary-literature (papers, specs, standards; leaf over commentary) · **Access date for all web sources:** 2026-08-23 · **Repo sources pinned at** `2ce929f` (inputs/PINNED.md).

**How to read the confidence tags.** `[leaf]` = I read the primary artifact's exact bytes this sitting (repo file, or the run cache under `cache/<sha>`). `[abstract-leaf]` = I read the paper's own abstract at the arXiv `/abs/` page (cached), not the full body. `[secondary]` = a search-surfaced summary of a primary I did not open at leaf this sitting — a candidate for the synthesizer to pin or drop. `[derived]` = my reasoning over leaf sources. `[FLAG]` = a figure or claim I deliberately do **not** assert as fact; provenance is too weak and it must be leaf-verified or dropped at synthesis.

This lane took **hypothesis Q2 first** (the in-repo structural decomposition — settled from the two design docs at leaf), then breadth across Q1/Q4/Q5 (external census, my lens' strong suit) and Q3 (the OKF standard read at leaf).

---

## Part A — Q2: are self-improve, dream, and memory one machine at three altitudes?

**Hypothesis under test (Q2):** the three in-repo loops are one background-loop kernel — `schedule → gather candidates → consolidate → corroboration/confidence-gated promotion → human-gated top rung → git commit` — differing only in payload and in which rung the human gates; if so, a single kernel is the borrowable shared mechanism. Contrary: the *consolidate* step differs in KIND, so they must stay separate.

**Grounding (all `[leaf]`):**
- `plans/memory-architecture.md` — the dream loop and OKF memory store. Source note: repo file at pin `2ce929f`; read whole.
- `plans/claude-port-plan.md` §3c and §5 — `/self-improve`, scheduling, guardrails, phases. Source note: repo file at pin `2ce929f`; read whole.
- `plugins/sleeper-service/README.md` — "Writes only to `research/` and `ideas/`; promotion into rules/skills requires the human." Source note: repo file at pin `2ce929f`; read whole.

### A.1 The shared skeleton is real (hypothesis partially CONFIRMED)

Decomposed against one step-vector, the three loops line up on five of six steps `[derived from the two leaf docs]`:

| Step | self-improve (`/self-improve`) | dream (`/dream`) | memory (OKF store lifecycle) |
|---|---|---|---|
| **schedule** | daily default, human-opt-in, manual always available (port-plan §3c, Resolved-decision 6) | "daily by default … reuses the port plan's existing scheduling story" (memory-arch §7.6) | n/a — memory is the *substrate* the other two write; it has no scheduler of its own |
| **gather candidates** | enumerate own rules/skills/agents/ideas/research → pick one (port-plan §3c) | collect `short-term/*` + new `MEMORY.md` auto-memory (memory-arch §7.5.1) | trajectory/`MEMORY.md` note → `short-term/<date>.md` (memory-arch §6 ladder rung 1) |
| **consolidate** | **run a FEOV research debate** — `/research "how should X evolve?"` (port-plan §3c) | **overlap-dedup merge** — `memory-consolidator` obeys expand-existing-before-append, sets `supersedes` (memory-arch §7.5.2) | **passive promotion ladder** — corroboration count `review_count ≥ 2` (memory-arch §6) |
| **gated promotion** | idea stub with alternatives; writes only `research/`+`ideas/` (port-plan §3c) | promote corroborated candidate to `status: active`, `confidence ≥ 0.7` (memory-arch §6) | `active.md → CLAUDE.md → skills/rules/` (memory-arch §6 ladder) |
| **human-gated top rung** | promotion into rules/skills requires the human — Semantic Consent (sleeper README; port-plan §3c guardrail) | promotion to a rule-skill "always human-gated" (memory-arch §6); optional PR for the project store (§7.5.5) | promote-to-global is "`/dream --promote-to-global` with human confirmation" (memory-arch §8) |
| **git commit** | git-tracked run dirs (port-plan Part 2) | "one git commit per store … so the whole pass is one reviewable diff" (memory-arch §7.5.5) | "Git is the substrate. Every memory mutation is a commit" (memory-arch §2, design goal 2) |

So the **outer frame is one machine**: schedule → gather → consolidate → confidence/corroboration-gated promotion → human-gated behavior change → commit. Four of the five findings the operator brief asks for (shared scheduling, promotion ladders, human gates, git as substrate) are **structurally shared, `[leaf]`-confirmed**. That is the borrowable "background-loop kernel."

### A.2 The consolidate step differs in KIND — the contrary finding also holds

The hypothesis' own disconfirmer is satisfied at the same time. The **consolidate** row above is three different mechanisms, not one:

- **self-improve consolidates by adversarial debate** — it *calls FEOV* (`/research`), so its consolidation is the whole blue/red/bench loop this report is about. `[leaf: port-plan §3c]`
- **dream consolidates by overlap-dedup** — a single `memory-consolidator` agent deciding expand-vs-append-vs-supersede on title/topic overlap. No adversary. `[leaf: memory-arch §7.5.2]`
- **memory "consolidates" passively** — it is a promotion *ladder* gated on a corroboration counter (`review_count`), not an active merge agent at all. `[leaf: memory-arch §6]`

**Reconciled finding (this is the honest Q2 answer, `[derived]`):** the three loops share an outer *kernel* (schedule/gather/promote/gate/commit) but diverge irreducibly at the *consolidate* core. A refactor should therefore extract the **kernel** (scheduling recipes, the promotion-ladder state machine, the human-gate-at-the-top-rung invariant, git-commit-per-pass) as shared machinery, while keeping *consolidate* a **pluggable strategy** with three implementations: `debate` (self-improve), `dedup-merge` (dream), `corroboration-count` (memory). This is a stronger, more actionable result than either pole of the hypothesis alone: it is neither "one machine, merge them" nor "three machines, keep them apart," but "one kernel, three consolidation strategies." Each half is filable as a tracked issue (extract-kernel; declare-consolidate-strategy-interface).

### A.3 A named coupling the docs already assert

self-improve and dream are not merely parallel — the port plan wires them together: `/self-improve` "**consumes the friction records** … as a first-class input" (port-plan §5, Phase 4). Friction is the FEOV complaint channel (research-protocol §Friction). So the self-improve loop's *gather* step is fed by the FEOV loop's exhaust. That is a real dependency edge, not a shared-kernel abstraction, and it belongs in the "what each borrows from the others" column: **self-improve borrows FEOV's friction stream as its candidate source.** `[leaf: port-plan §5]`

---

## Part B — Q1 & Q5: verification LOCUS (who owns the gate) and FORM (is it re-runnable)

These two hypotheses are orthogonal (Q5 says so explicitly) and the primary literature separates cleanly along both axes. My census of external systems, each verified at the `/abs/` leaf or flagged secondary:

| System | Verification LOCUS (Q1) | Verification FORM (Q5) | Human gate (Q4) | Source note |
|---|---|---|---|---|
| **Self-Refine** (Madaan et al. 2023) | **self** — "the same pre-trained model acts as the initial answer generator, the feedback provider, and the refiner" | prose self-critique (no execution) | none (a technique, not a deployed loop) | `[secondary]` github.com/madaan/self-refine; paper "Self-Refine: Iterative Refinement with Self-Feedback", arXiv:2303.17651 |
| **Reflexion** (Shinn et al. 2023) | **mixed** — self-reflection over feedback that is "scalar or free-form" and from sources "external or internally simulated" | often EXECUTION on coding/RL tasks (env reward, unit tests) feeds the reflection | none | `[abstract-leaf via search]` NeurIPS 2023, proceedings.neurips.cc/paper_files/paper/2023/file/1b44b878bb782e6954cd888628510e90-Paper-Conference.pdf; repo github.com/noahshinn/reflexion |
| **AI-Scientist** (Lu et al. 2024) | **self / same-family** — "runs a simulated review process"; "an automated reviewer … achieves near-human performance" | executes experiments, BUT the *paper-quality* gate is the LLM reviewer, not a ground-truth check on the claims | none in the loop (open-ended iteration) | `[abstract-leaf]` cache sha `f22d5ff…`, arXiv:2408.06292 |
| **Darwin Gödel Machine** (Zhang et al. 2025) | **external ground-truth** — "empirically validates each change using coding benchmarks" (SWE-bench 20→50%, Polyglot 14.2→30.7%) | EXECUTION (benchmark test suites) | architecturally none; **experiments run "with safety precautions (e.g., sandboxing, human oversight)"** | `[abstract-leaf]` cache sha `30ff41b…`, arXiv:2505.22954 |
| **AlphaEvolve** (Novikov et al. 2025) | **external ground-truth** — "continuously receiving feedback from one or more evaluators"; results "provably correct" | EXECUTION + automatic evaluation ("grounded using code execution and automatic evaluation") | none in the loop (evaluator is the gate) | `[abstract-leaf]` cache sha `7a0d9ac…`, arXiv:2506.13131; DeepMind blog |
| **FunSearch** (Romera-Paredes et al. 2024) | **external ground-truth** — an `evaluate` function scores candidates; kept "if correct" | EXECUTION (the evaluator runs the program) | none in the loop | `[secondary]` Nature s41586-023-06924-6 |
| **Deep-research products** (OpenAI / Gemini) | model-internal synthesis; no adversarial gate | prose synthesis of retrieved sources; no execution gate | **at the QUERY** — clarifying questions before the loop; time-boxed (Gemini API hard 60-min cap); single-run deliverable | `[secondary]` bytebytego / MindStudio Gemini Deep Research API; arXiv:2506.12594 survey |

### B.1 Q1 finding — the external-adversary claim, corrected at the leaf

The round-0 hypothesis put DGM in the "self-grade" camp ("benchmark self-selection"). **The leaf refutes that placement `[abstract-leaf]`:** DGM's *fitness* is external benchmark execution, not self-review — its only self-driven step is the archive *selection* heuristic ("interestingly new"), which chooses what to mutate, not whether it worked. DGM therefore belongs with AlphaEvolve/FunSearch in the **execution-verified** camp, not with Self-Refine/AI-Scientist in the **self-review** camp. Carrying this correction is worth more than the tidy original grouping.

So the corpus splits three ways on LOCUS, not two:
1. **Self / same-family review** — Self-Refine, AI-Scientist's reviewer. This is where the documented failure lives.
2. **External ground-truth by execution** — DGM, AlphaEvolve, FunSearch. Reliable *because* the verifier runs code, but only applicable where a machine-checkable objective exists (a benchmark, an `evaluate` function).
3. **No adversarial gate at all** — deep-research products synthesize and hand off.

**Where FEOV sits, and the disconfirmer that fires `[derived]`:** the round-0 hypothesis was that an externalized adversary is a genuine differentiator. The disconfirming search (see Lines of Inquiry, pursued) shows adversarial/separate-critic patterns are **actively emerging in 2025-26** — e.g. CMIP-Forge's "Autonomous Adversarial Peer Review Protocol" subjecting worker analyses to independent multi-model critique (arXiv:2606.17076 `[secondary]`), and a broader generator-verifier literature arguing a model checking its own work is capped at its own generation accuracy while a separate adversary breaks the symmetry (arXiv:2602.13213, and the self-preference-bias result arXiv:2410.21819 `[secondary]`). **So "has an external adversary" is NOT unique in kind, and the claim must narrow.** The defensible FEOV differentiator is the *specific combination*, none of whose parts I found co-located in the external corpus:
- the adversary **owns a binary PASS gate it is chartered to withhold** (red owns PASS/FAIL — port-plan §3b, research-protocol), rather than critique that the generator may revise-and-ignore;
- a **separate bench owns termination** ("is it close enough" — research-protocol §Termination), so *neither* generator nor adversary declares the loop done;
- every exchange is a **refusable record event** read back through projections (research-protocol §"The exchange is TOOL-MEDIATED"), not prose passed between agents;
- **computation-gaps are uncloseable on prose** (blue-researcher constitution: `awaiting_proof`), importing execution-verification into open-ended *research* rather than only code.

That is the accurate, leaf-survivable version: FEOV **formalizes and combines** an emerging pattern; it does not invent the adversary. Stating it larger than that would not survive red.

### B.2 Q5 finding — execution as the countermeasure, precisely scoped

The failure mode the operator brief names ("plausible-but-wrong papers") maps exactly onto camp 1. AI-Scientist executes its *experiments* but gates *paper quality* with an LLM reviewer `[abstract-leaf]`; the self-preference-bias and generator-verifier literature explains why that gate is structurally weak `[secondary]`. Camps 2's reliability comes from the evaluator EXECUTING — this is `[abstract-leaf]`-confirmed for DGM and AlphaEvolve and `[secondary]` for FunSearch. **FEOV's `proof:` mechanism (rerunnable script + output, computation-gaps uncloseable on prose) is camp-2's execution-verification imported into the OPEN-ENDED research task class**, where camps 1 and 3 mostly leave verification to prose. The disconfirmer (does any peer already mandate execution-verification for open-ended research?) did not fire against the strong form: the execution-verified systems all operate on *closed* objectives (benchmarks, `evaluate` functions), not open research prose. FEOV's move is applying it where the objective is not machine-given — which is exactly why it needs a human-legible adversary and a bench on top of the executable proofs.

---

## Part C — Q4: WHERE the human gate sits (the risk-posture axis)

Ordering the corpus by gate position `[derived from the Q1/Q5 census]`:

- **No gate in the loop** — DGM (architecturally), AlphaEvolve, FunSearch. The automated evaluator IS the gate; self-modification/discovery proceeds unattended. **Caveat, `[abstract-leaf]`:** DGM's *paper* reports its experiments were run "with safety precautions (e.g., sandboxing, human oversight)" — so "no human gate" is a claim about the *architecture*, not about what was actually run. This corrects the round-0 phrasing that DGM "places NO human gate."
- **Gate at the query** — deep-research products: the human scopes the question up front, then the system runs to a one-off deliverable and persists nothing behavioral `[secondary]`. (Persistence-across-runs I could not confirm at leaf; stated as product-dependent, not "never persists.")
- **Gate uniquely late and narrow** — sleeper-service/FEOV: the loop "writes only `research/` and `ideas/`" and gates *promotion into behavior-changing rules/skills* on the human `[leaf: sleeper README; port-plan §3c]`. Dream gates the same rung (promote-to-rule-skill, promote-to-global — `[leaf: memory-arch §6, §8]`); memory gates promotion-to-global `[leaf: memory-arch §8]`.

**Q4 finding, `[derived]`:** the three in-repo loops **converge on gate position** — *propose-freely, human-ratifies-behavior-change* — while diverging on payload (rules / consolidated knowledge / substrate concepts). This "propose-but-never-self-modify-behavior" posture is the shared safety invariant, and it is the axis that most cleanly separates the in-repo loops from the fully-closed self-improvers (DGM/AlphaEvolve): same outer loop, opposite gate placement. The disconfirmer (an external system gating at exactly this rung) did not fire — the execution-verified systems gate on the evaluator, not on human ratification of a behavior change; the deep-research products gate earlier (at the query) and don't self-modify at all. Q4 is the hinge: it is the one axis on which FEOV's in-repo family is genuinely distinctive against the external corpus.

---

## Part D — Q3: the mediated-record novelty, and the OKF standard read at leaf

**Hypothesis (Q3):** the load-bearing novelty of the FEOV loop is not multi-agent debate (common) but that every exchange is a *validated event on an append-only record read back through projections*, with the plausible-zero failure designed out; measured against that bar the dream/memory design (markdown+YAML consolidated by a merge agent, committed to git) is a "document standing in for a record" and structurally weaker than the loop it parallels. Disconfirmer: if OKF frontmatter is validated on write, the weakness narrows.

### D.1 "Debate is common" — confirmed

AutoGen (Wu et al. 2023, arXiv:2308.08155 `[secondary]`), CAMEL (Li et al. 2023 `[secondary]`), and multi-agent debate (Du et al. 2023, arXiv:2305.14325 `[secondary]`) establish that multi-agent conversation/debate is a well-populated design space. So Q3's premise holds: multi-agent *debate* is not where FEOV's novelty can honestly be located.

### D.2 The disconfirmer, tested at the OKF leaf — and it does NOT fire the way round-0 framed it

I read the OKF v0.2 spec at the leaf (`cache/26aa5da…`, github.com/GoogleCloudPlatform/knowledge-catalog/okf/SPEC.md). Two findings, and they pull in opposite directions:

**(i) OKF is permissive BY DESIGN — the facts-are-fields critique of the memory store STANDS `[leaf]`.** OKF §11 Conformance is explicitly non-refusing: consumers "MUST NOT reject a concept for missing any optional family," "MUST tolerate unknown types," "MUST tolerate broken links." The only hard conformance checks are "YAML parses" and "`type` is non-empty." The memory-architecture profile's lifecycle fields (`confidence`, `status`, `review_count`, `last_seen`, `supersedes`) are **all optional and recovered by reading, never refused at write** — exactly the "pattern standing in for a schema" shape facts-are-fields names. A malformed `confidence`, or a consolidator that stops writing `review_count`, fails *silently and looks like a clean board* — the plausible-zero. So the round-0 disconfirmer ("if OKF validates on write, the weakness narrows") **does not fire**: OKF v0.2 does not validate on write in the refusable sense, and the memory store's persistence layer is therefore, as designed, weaker than the FEOV record it parallels. Q3's core claim survives contact with the standard. `[leaf, high confidence]`

**(ii) But OKF v0.2 §10 "Attested Computation" is FEOV's `proof:` mechanism, independently invented `[leaf, and this is the most valuable finding in Part D].`** The spec defines a concept type carrying: a sanctioned computation; a **parameter-only surface** where "the agent MAY only supply *values* … it MUST NOT author or edit the computation"; an **executor** that runs it and returns a **receipt**; and a **deterministic (no-LLM) attester** that inspects the receipt and returns a verdict, so that "did the sanctioned thing run" is "a mechanical comparison rather than a judgement call" (OKF §10.2, §10.3). §10.6 distinguishes `verified` (doc-level, definition still matches policy) from *attestation* (a single run produced the value the sanctioned way) — the same distinction the blue-researcher constitution draws between a cited source and "a computation you ran and recorded." **This is convergent evolution:** Google's knowledge-format standard and FEOV independently arrived at "a computed fact must carry a re-runnable, mechanically-checkable receipt, not prose." It is direct, external, standards-body corroboration of the operator's standing requirement (proof, not prose).

### D.3 The actionable Q3 result — two filable findings `[derived]`

1. **The memory design is pinned to the wrong OKF version.** memory-architecture.md §3.1 pins to "OKF v0.1" and §9-item-7 says "OKF is v0.1 and evolving. We pin to a documented profile." **OKF v0.2 exists and already ships as first-class exactly what the profile hand-rolled**: `provenance`/`sources`, trust (`generated`/`verified` + trust tiers), lifecycle (`status`/`stale_after`), and the actor convention `[leaf: OKF §5, §13]`. The bespoke profile is now partly redundant with upstream. This is a live-source-drift finding: **re-pin the memory profile to OKF v0.2 and drop the hand-rolled fields it now duplicates.** (Publication dates v0.1≈2026-06-12 / v0.2≈2026-07-25 are `[secondary]`, startuphub/marktechpost; the spec body's "Version 0.2" is `[leaf]`.)
2. **Adopt OKF Attested Computation for any quantitative memory, and adopt the FEOV record's refusability for the consolidate step.** The memory store cannot make free-form knowledge concepts refusable-on-write (OKF forbids it), but it CAN (a) route any *computed/quantitative* memory through OKF §10 attestation, closing the plausible-zero for the numbers that matter most, and (b) borrow FEOV's mediated-record discipline for the dream loop's consolidation events (which today are a merge agent writing markdown), so a dropped/ malformed consolidation surfaces as a refusal rather than an empty board. **Q3's highest-value conclusion holds: the memory design should borrow FEOV's *record discipline*, not merely its *cadence*.**

---

## Lines of Inquiry — pursued, deferred, and rejected

**Pursued (this lane):**
- **Q2 structural decomposition** — settled from the two design docs at leaf; result is "one kernel, three consolidation strategies" (Part A), not either pole of the hypothesis.
- **Q1/Q4/Q5 external census via primary literature** — leaf abstracts for AI-Scientist, DGM, AlphaEvolve; secondary for Reflexion/Self-Refine/FunSearch/deep-research. Corrected DGM's camp (execution-verified, not self-grade) and DGM's gate ("no gate" is architectural, experiments had human oversight). Narrowed the FEOV differentiator to a specific combination rather than "has an adversary."
- **Q3 via the OKF standard at leaf** — read OKF v0.2 SPEC.md whole; found the permissiveness that confirms the critique AND the Attested-Computation convergence that corroborates the proof-discipline.
- **Disconfirming line (≥1-in-5 budget):** "do external systems already use adversarial/external critics and execution-verification?" — pursued via the generator-verifier / self-preference-bias / CMIP-Forge searches. **Outcome:** it fired against the *strong* novelty claim (adversary is emerging elsewhere) and forced the narrowing in B.1; it did NOT fire against the *combination* claim or the open-ended-research scoping in B.2. This is the most consequential thing the lane did — it moved the finding from "FEOV's adversary is novel" to "FEOV combines an emerging pattern in a way I could not find co-located."

**Deferred (worth a later run, not this lane):**
- **Leaf-verify the generator-verifier "≈30% self-review vs ≈85% adversarial" figure `[FLAG]`.** It appears in a vendor blog (augmentcode.com) surfaced by search and is directionally echoed by the self-preference-bias paper, but I did not reach a primary at leaf. I deliberately do NOT assert the numbers. A later round should fetch arXiv:2602.13213 / the underlying study and either pin the figure to a primary or drop it; the qualitative direction is independently supported and stands without it.
- **Full-text leaf reads of AI-Scientist v2 (arXiv:2504.08066) and the DGM body** for the exact failure-mode and human-oversight passages, if red challenges the abstract-level grounding.
- **OpenAI/Gemini deep-research persistence** — confirm at leaf whether these products persist anything across runs, to firm up the Q4 "never persist" clause I softened to "product-dependent."

**Considered and rejected (so a later run does not re-walk them):**
- **Write a program to formally align the three loops' step-vectors (a computed proof for Q2).** Rejected: the alignment is definitional/structural, not arithmetic — there is no rate to sum or enumerate that a table does not already settle. A program here would be theater, not evidence. The step-vector table in A.1 is the honest artifact; forcing a `proof:` would violate repair-minimalism.
- **Treat the whole topic as an empirical literature question and skip the in-repo structural read.** Rejected on the constitution's "both avenues" clause: Q2 is settled by *reading the two design docs*, and the biggest single finding (D.2, OKF≈FEOV proof convergence) came from reading a *standard* at leaf, not from counting citations. Classifying the question as "empirical" up front would have missed both.

---

## Friction note (for the envelope, not the report)
The engine's PreToolUse hook is not injecting `--run` for this run (surfaced at register); every seat is affected. Recorded via the friction verb. No capability blocked the research itself — the `fetch` verb served leaf bytes cleanly and the cache made them re-readable.
