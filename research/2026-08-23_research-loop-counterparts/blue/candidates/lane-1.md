# Lane 1 candidate draft — adversarial-disconfirming-first

**Seat:** blue-lane-1 · **Method lens:** adversarial-disconfirming-first (hunt evidence AGAINST
the frontier hypotheses before evidence for them). **Order:** hypothesis Q1 first, then breadth
across Q4/Q5 (external axis) and Q2/Q3 (in-repo axis).

**How to read this draft.** Every claim about an external system is tagged with the state red
cares about: `[leaf-verified]` (I read the primary abstract/source bytes), `[search-digest]` (a
search engine's summary of the source, not yet leaf-read — a synthesizer must leaf-verify before
citing), or `[design-doc]` (read at the leaf from the pinned in-repo design docs at commit
`2ce929f`). Numbers carry their provenance inline. The lens is disconfirming, so the spine of each
section is "here is the evidence that the hypothesis is WRONG or overclaimed," and the surviving
claim is whatever is left after that assault.

---

## Q1 — Is FEOV's externalized adversary a genuine differentiator, or table stakes?

**Frontier hypothesis (on the record):** external automated-research/self-improving systems
predominantly SELF-grade (Reflexion self-reflection, Self-Refine self-feedback, AI-Scientist LLM
reviewer, Darwin-Gödel-Machine benchmark self-selection); FEOV's structural differentiator is a
gate owned by a SEPARATE ADVERSARIAL seat (red) issuing a binary PASS it is chartered to withhold,
plus a bench judging termination. Disconfirmer: if independent/adversarial critics OR ground-truth
execution are already the norm, the differentiator collapses.

### Disconfirming evidence found (the assault)

**D1 — Separate critic models guided by an explicit constitution are already a paradigm, not a
novelty.** Constitutional AI / RLAIF replaces the human rater with a "constitution-following critic
model" that provides preference labels; the critique-and-revise loop is guided by an explicit set
of principles. This is structurally the same shape as "red, a critic bound to a written
constitution." [search-digest — "Constitutional AI: Harmlessness from AI Feedback", arXiv
2212.08073, https://arxiv.org/pdf/2212.08073 , accessed 2026-08-23; corroborated by Repello AI
glossary, https://repello.ai/glossary/constitutional-ai , accessed 2026-08-23] The differentiator
is therefore NOT "a critic bound by written principles" — that exists. What is left: CAI's critic
emits *training-signal preference labels*, it does not own a binary gate that blocks a *deliverable*
from assembly, and it is not chartered to WITHHOLD across rounds. The FEOV distinction narrows from
"externalized constitution-bound critic" to "externalized critic that GATES the artifact."

**D2 — Adversarial reviewers explicitly calibrated to REJECT AI-written work already exist, and one
2026 system names FEOV's exact rationale.** The Red Queen Gödel Machine (RQGM, Cambridge/NVIDIA,
June 2026) makes "evaluation part of the improvement loop and opening search to evolving evaluators,
adversarial objectives"; it observes that "the strongest baseline reviewer over-accepts
AI-generated papers at up to 1.91x the human rate" and "corrects this by introducing an adversarial
objective that discovers reviewers equally stringent on AI and human work." It adds an
"agent-as-a-judge code-review signal" and reports co-evolved graders reaching 9% higher
ground-truth accuracy. [leaf-verified — abstract of "The Red Queen Gödel Machine: Co-Evolving Agents
and Their Evaluators", arXiv 2606.26294, https://arxiv.org/abs/2606.26294 , accessed 2026-08-23]
This is the single strongest disconfirmer: an externalized adversarial evaluator, explicitly
motivated by escaping self-preference bias, is in the literature. FEOV's "red owns the gate to
defeat self-grading" is therefore a member of a recognized class, not a lone invention.

**D3 — Agentic judges that read the whole trajectory (not just the final output) and carry
tool-use + memory are an established evaluation paradigm.** Agent-as-a-Judge (Meta/KAUST, Oct 2024)
"equips the AI evaluator with agent-like capabilities such as tool use, memory, and multi-step
reasoning, so the judge is itself an autonomous agent that can observe and critique each step."
[search-digest — "Agent-as-a-Judge: Evaluate Agents with Agents", arXiv 2410.10934,
https://arxiv.org/pdf/2410.10934 , accessed 2026-08-23; ICML 2025 poster, https://icml.cc/virtual/2025/poster/45485]
This is precisely FEOV's red re-reading the whole living report in context with `memory: project`.
So "the auditor reads the entire artifact and remembers past patterns" is not distinctive either.

**D4 — AI-Scientist DOES have a distinct reviewer role, so "self-grade" understates it.** The
AI-Scientist "runs a simulated review process for evaluation … we design and validate an automated
reviewer, which we show achieves near-human performance in evaluating paper scores," producing a
NeurIPS-guideline binary accept/reject. [leaf-verified — abstract of "The AI Scientist", arXiv
2408.06292, https://arxiv.org/abs/2408.06292 , accessed 2026-08-23; reviewer detail corroborated by
Towards AI newsletter #113, https://newsletter.towardsai.net/p/113-sakanas-ai-scientist ,
accessed 2026-08-23] The reviewer is a *separate agent role*. The hypothesis's framing of
AI-Scientist as pure "self-grade" is imprecise — it has a reviewer. What makes it SELF-grading is
that the reviewer is the same model family serving the generator, and the loop's own paper is judged
by the loop's own reviewer; the abstract itself closes the circle: "papers that exceed the
acceptance threshold at a top machine learning conference *as judged by our automated reviewer*."

### What survives the assault (the corrected, narrower claim)

The externalized-adversary framing survives only in a NARROWED, COMBINATION form. No single axis is
unique:
- constitution-bound critic → exists (CAI/RLAIF, D1)
- adversarial reviewer calibrated against AI-preference → exists (RQGM, D2)
- whole-trajectory agentic judge with memory → exists (Agent-as-a-Judge, D3)
- distinct reviewer role in a research loop → exists (AI-Scientist, D4)

What is not attested together in any single external system I found: a critic that (a) owns a
**binary PASS gate blocking the deliverable's assembly**, (b) is chartered to **withhold** rather
than to score-and-average, (c) is a **fixed constitution-bound seat** (not co-evolved, not a training
signal), (d) writes/reads exclusively through a **refusable append-only record**, and (e) sits in
front of a **human-gated promotion** into behavior. FEOV's differentiator is this *bundle*, and the
honest headline is "FEOV composes known adversarial-evaluation primitives into a
deliverable-gating configuration," NOT "FEOV invents adversarial external verification."

**Calibration:** high confidence that no single axis is novel (D1–D4 are leaf- or
digest-verified against named systems); medium confidence on the "bundle is unattested elsewhere"
claim — it rests on absence of a counterexample in ~15 searches, which is weaker than a positive
find. A synthesizer should mark the bundle claim as "not contradicted by the corpus surveyed,"
not "proven unique."

### A finding that cuts AGAINST FEOV (the disconfirming lens turned on the loop itself)

**The self-preference-bias literature partially undercuts FEOV's own escape route.** Self-preference
bias in LLM-as-judge is real and large — on ArenaHard the bias ranges from -38% to +90%, and
critically, "ensembling across judge models address[es] variance but not the systematic biases
shared across judges." [search-digest — "Self-Preference Bias in LLM-as-a-Judge", arXiv 2410.21819,
https://arxiv.org/abs/2410.21819 , accessed 2026-08-23] And across FAMILIES: "models sharing
architecture and training lineage tend to exhibit higher internal agreement … multi-judge consensus
across different model families … reduce but do not eliminate this bias." [search-digest —
"Reliability without Validity: … LLM-as-a-Judge …", arXiv 2606.19544,
https://arxiv.org/html/2606.19544v1 , accessed 2026-08-23] This run configures blue as `fable` and
red/bench as `sonnet` (run-config.json) — different models, but plausibly shared training lineage
(both Anthropic). So FEOV's externalization *reduces but does not eliminate* the shared-bias failure
its own design invokes to justify red. This is a genuine limit red should force blue to state, not a
fatal flaw: the human promotion gate and execution-proof anchors are the backstops that a pure
LLM-judge lacks.

---

## Q5 — Is execution-anchored proof FEOV's targeted countermeasure, or already standard?

**Frontier hypothesis:** the documented weakness of automated-research systems is verification a
skeptic cannot re-run (AI-Scientist plausible-but-wrong papers, LLM-as-judge unreliability); FEOV's
`proof:` anchor (rerunnable script + output; computation-gaps uncloseable on prose) is the targeted
countermeasure; prediction: the reliable systems are those whose evaluator EXECUTES.

### Disconfirming evidence (the assault)

**D5 — Execution-as-ground-truth is the NORM in program-search systems, not a FEOV import.**
FunSearch/AlphaEvolve pair an LLM with an automated evaluator: "programs are then automatically
executed and assessed … every new algorithm is grounded by automated evaluation through code
execution … completely avoiding the risk of LLM hallucinations." [search-digest — AlphaEvolve
coverage, https://www.getmaxim.ai/blog/alphaevolve-ai-for-scientific-discovery/ , accessed
2026-08-23; FunSearch mechanism corroborated by R.C. Suwandi blog,
https://richardcsuwandi.github.io/blog/2025/llm-algorithm-discovery/ , accessed 2026-08-23] The
Darwin-Gödel Machine "empirically validates each change using coding benchmarks" (SWE-bench Verified
20.0%→50.0%, Polyglot 14.2%→30.7% over 80 iterations). [search-digest — "Darwin Gödel Machine",
arXiv 2505.22954, https://arxiv.org/abs/2505.22954 , and Sakana https://sakana.ai/dgm/ , accessed
2026-08-23] Voyager saves a skill only as a "verified code snippet," verified by environment
execution. [search-digest — Voyager, https://aiunderstanding.org/learn/voyager-and-skill-library-agents ,
accessed 2026-08-23] The prediction "reliable systems execute" is CONFIRMED — but that means
execution-verification is table stakes in the narrow-objective world, so FEOV cannot claim to have
invented it.

**D6 — The theoretical ancestor already gates change on PROOF.** Schmidhuber's Gödel Machine
rewrites its code only "as soon as it has found a proof that the rewrite is useful," and "must ignore
those self-improvements whose effectiveness it cannot prove." [leaf-verified via IDSIA page —
"Gödel Machines: Self-Referential Universal Problem Solvers", https://people.idsia.ch/~juergen/gmweb2/gmweb2.html
and arXiv cs/0309048, https://arxiv.org/pdf/cs/0309048 , accessed 2026-08-23] "Gate change on a
proof you can check" is a 2003 idea. FEOV's `proof:` anchor is a *pragmatic weakening* of it (a
rerunnable script standing in for a formal proof), which is a virtue, not a novelty.

### What survives (narrowed claim)

The defensible claim is narrow and specific: **execution-anchored verification is standard for
narrow objective-function tasks (program search, coding benchmarks, game environments) but is
largely ABSENT from OPEN-ENDED research-report generation, where AI-Scientist/STORM/deep-research
products fall back on model review or source-corroboration.** FEOV's move is to *require* a
rerunnable proof for the computational sub-claims *inside* an open-ended research artifact, and to
make a computation-gap uncloseable on prose. That transplant — execution ground-truth into the
open-ended-research task class — is the honest contribution. The failure it targets is
leaf-verified real: AI-Scientist-class output exhibits "faked experimental results … hallucinated
methodology … incorrect citations … mathematical errors," and NeurIPS 2025 had ≥100 confirmed
hallucinated citations across 53 accepted papers (~1%) *despite* peer review. [search-digest —
Fortune, https://fortune.com/2026/01/21/neurips-ai-conferences-research-papers-hallucinations/ , and
Nature, https://www.nature.com/articles/d41586-026-00969-z , accessed 2026-08-23; failure taxonomy
from "MLR-Bench", arXiv 2505.19955, https://arxiv.org/pdf/2505.19955 , accessed 2026-08-23]

**Distinction from deep-research verification (a boundary red will probe):** OpenAI/Gemini deep
research do "multi-level verification … across multiple independent sources" and use dedicated
citation agents so "every assertion … can be traced back to a verified source." [search-digest —
OpenAI "Introducing deep research", https://openai.com/index/introducing-deep-research/ , and
ByteByteGo, https://blog.bytebytego.com/p/how-openai-gemini-and-claude-use , accessed 2026-08-23]
That is *source-corroboration*, a different mechanism from *re-runnable execution*. FEOV has both
(cite for corroboration, proof for computation); the claim must not conflate them.

---

## Q4 — Is the human gate's PLACEMENT the distinctive risk-posture axis?

**Frontier hypothesis:** gate position separates these systems by risk posture; fully-closed
self-improvers (DGM, AlphaEvolve) place NO human gate; deep-research products gate at the query and
never persist; sleeper-service/FEOV place it uniquely late and narrow (writes only research/+ideas/,
gates promotion into behavior-changing rules/skills) — the same rung dream and memory gate.

### Evidence: the axis holds, with one disconfirmer

The census across the systems above, arranged by gate position, is coherent:

| System | Verification locus | Human gate position | Persists to behavior? |
|---|---|---|---|
| Gödel Machine / DGM / AlphaEvolve / FunSearch | execution / proof | NONE (closed loop) | yes — rewrites own code / weights-free skill |
| Voyager | environment execution | NONE | yes — skill library |
| AI-Scientist(-v2) | self-review (LLM/VLM reviewer) | only at final submission | no (produces papers) |
| STORM | none (retrieval + write) | NONE | no |
| OpenAI/Gemini deep research | source-corroboration | at the QUERY (input), advisory on output | no (never persists) |
| **sleeper-service / FEOV** | adversarial red + execution proof | at PROMOTION into rules/skills | yes, but human-gated |

[table synthesized from the sources cited in Q1/Q5 above, all accessed 2026-08-23]

**D7 — the disconfirmer: "human-approval gate at promotion" is an emerging safety pattern, not a
FEOV monopoly.** A security-hardened self-improvement skill for OpenClaw ships "mandatory
human-approval gate, automated sanitization, audit tooling, and promotion rate-limiting."
[search-digest — gateswell/safe-self-improvement-agent,
https://github.com/gateswell/safe-self-improvement-agent , accessed 2026-08-23] And the broader
principle "a self-improvement change rolls forward only when a held-out evaluation the optimizer
can't see improves … the contract behind eval-gated online learning" is stated as a general safety
principle. [search-digest — "Self-Improving AI: What Actually Works in 2026",
https://www.morphllm.com/self-improving-ai , accessed 2026-08-23] So the placement axis is real and
FEOV sits at the conservative end, but "gate at promotion" is a converging community pattern, not a
unique posture. FEOV's specific narrowing — the loop's *write surface* is physically confined to
`research/` and `ideas/` (port-plan §3c guardrail) so the gate cannot be bypassed by the loop
writing behavior directly — is the sharper, verifiable distinction. [design-doc — claude-port-plan.md
§3c, commit 2ce929f]

---

## Q2 — Are self-improve, dream, and memory "one machine at three altitudes"?

**Frontier hypothesis:** the three in-repo loops are one machine — each is (schedule → gather
candidates → consolidate → corroboration/confidence-gated promotion → human-gated top rung → git
commit), differing only in payload; a single background-loop kernel is the borrowable shared
mechanism. Contrary: the consolidate step differs in KIND (dedup-merge vs adversarial debate vs
passive ladder), so they must stay separate.

### Structural decomposition at the leaf (design docs, commit 2ce929f)

| Loop | Schedule | Gather | **Consolidate (the core)** | Promotion gate | Substrate |
|---|---|---|---|---|---|
| **self-improve** (port-plan §3c) | daily default, `claude -p` | enumerate own rules/skills/agents/ideas/research | **run a full FEOV adversarial research debate** (`/research "how should X evolve?"`) → idea stub w/ alternatives | human at promotion to rules/skills (Semantic Consent) | git; writes only research/+ideas/ |
| **dream** (memory-arch §7.5) | daily default, `claude -p` | short-term notes + new MEMORY.md | **overlap-dedup MERGE** by a `memory-consolidator` agent (expand-existing-before-append; set supersedes; bump review_count/confidence) | human ratifies via optional PR on the project store; rule-skill promotion human-gated | git; one commit per pass |
| **memory** (memory-arch §6) | (substrate, driven BY dream) | trajectory → short-term | **passive promotion LADDER** gated on review_count≥2 (corroboration) + confidence≥0.7 | human at top rung (promote-to-global, rule-skill) | git |

**Finding (both hypotheses partly true; the honest result is a SPLIT):** the OUTER loop is genuinely
shared — schedule → gather → [core] → confidence/corroboration-gated promotion → human-gated top
rung → git commit. This is a real borrowable kernel: a "scheduled, git-native, human-gated-at-
promotion background-loop harness." But the CONSOLIDATE core differs **in kind**, exactly as the
contrary predicted: an adversarial multi-agent debate (self-improve) is not the same machine as an
overlap-dedup merge (dream) or a passive corroboration ladder (memory). Therefore the actionable
recommendation is a SPLIT, not a unification: **factor out the outer harness as shared
infrastructure; keep the three consolidation engines separate and documented as intentionally
distinct.** Collapsing them into one "kernel" that also unifies the core would force a debate engine
and a dedup merge into one abstraction that fits neither — a design made worse to satisfy a symmetry
that the evidence does not support. [design-doc — memory-architecture.md §6, §7.5, §7.6;
claude-port-plan.md §3c, commit 2ce929f]

**Calibration:** high confidence on the split, because it is a direct structural read of the two
design docs at the pinned commit, not an inference. The one caveat a synthesizer must carry: these
are DESIGNS (memory-architecture.md is "Status: Proposal"; sleeper-service README says "scaffold
only (Phase 0)"), so this compares *intended* machinery, not shipped behavior. That belongs in every
sentence that states the finding.

---

## Q3 — Is the dream/memory store a "document standing in for a record," structurally weaker than
FEOV's refusable record?

**Frontier hypothesis:** the load-bearing novelty of the FEOV loop is that every exchange is a
VALIDATED event on an append-only record read back through projections, with the plausible-zero
failure designed out; measured against that bar, the dream/memory architecture as designed
(markdown+YAML consolidated by a merge agent, committed to git) is a document standing in for a
record and is structurally WEAKER than the loop it parallels. Disconfirmer: if OKF frontmatter is
validated on write, the weakness narrows.

### Testing the disconfirmer at the leaf

The disconfirmer FAILS — i.e. the hypothesis largely HOLDS — on a direct read of the design:
- The OKF profile makes "only `type` … strictly required (OKF compliance); the lifecycle fields
  (`status`, `confidence`, `last_seen`, `review_count`, `supersedes`, `provenance`) are our
  profile's additions"; "A file missing them is still a valid OKF concept — it is simply treated as
  a low-confidence candidate." [design-doc — memory-architecture.md §3.1, commit 2ce929f] So there
  is NO write-refusing schema: a malformed or field-absent concept is silently accepted as
  low-confidence, which is exactly the "plausible zero reads like a clean board" failure
  facts-are-fields names.
- The design ITSELF concedes the failure mode: §9.4 "Bad merges silently lose knowledge (the
  OpenClaw diary's 'details unavailable' failure)"; §2 cites the OpenClaw Dream Diary degrading to
  "a memory trace surfaced, but details were unavailable" as "a cautionary tale about consolidation
  with no schema." [design-doc — memory-architecture.md §2, §9.4, commit 2ce929f]
- The only "refuse" in the dream/memory loop is the OPTIONAL human PR review on the project store
  (§7.5 "Optionally open a PR … so a human ratifies") and `git revert` as undo — a human gate and a
  reversal, not a machine that refuses a malformed write at the point of writing.

Contrast with FEOV: `feov-record` is "a verb that can refuse"; "the record is the read path and the
.md files are for human verification"; the whole design is built so a miss is LOUD, not a plausible
zero. [design-doc — research-protocol SKILL.md "The exchange is TOOL-MEDIATED"; facts-are-fields
SKILL.md, commit 2ce929f]

### Disconfirming the disconfirming lens — the honest counterweight (facts-are-fields' OWN brake)

facts-are-fields explicitly scopes itself NARROW: "This is about bypassing a record you already
have … The rule does NOT say 'any two things sharing a string need a schema between them'." And:
"Where no record exists, creating one is a design decision with a cost, not an obligation." So the
finding must NOT become "the memory store is broken because it isn't feov-record." The memory
store's job (accumulate consolidated knowledge for human+agent reading, versioned in git) is
different from the debate record's job (adjudicate a live adversarial exchange where a plausible
zero silently passes a bad claim). The correct, cost-aware finding: **the dream/memory persistence
layer has an unmediated-write weakness the design already half-acknowledges; the borrowable
mitigation from FEOV is not "adopt feov-record" but "add a write-time validator that REFUSES a
concept missing its load-bearing lifecycle fields, so an absent field is an error the author sees,
not a low-confidence silent accept."** That is a small, bounded control (a pre-commit or
consolidator-side schema check), not a re-architecture — and it is exactly what facts-are-fields
recommends when a record already exists ("put it in a field a writer can refuse"). This is a
concrete, file-able convergence finding.

**External corroboration that this is a live design axis, not a repo idiosyncrasy:** 2026
provenance/auditability research treats validated, append-only agent records as the emerging bar —
"Clarus … makes provenance an active collaboration substrate, where intermediate artifacts, failed
attempts, tool executions, decision records … are captured as audit checkpoints." [search-digest —
"Clarus", arXiv 2606.30246, https://arxiv.org/pdf/2606.30246 , accessed 2026-08-23] and
"Claim-Level Auditability for Deep Research Agents" [search-digest — arXiv 2602.13855,
https://arxiv.org/html/2602.13855 , accessed 2026-08-23]. So feov-record's refusable-record
discipline is FEOV's instance of a recognized 2026 class — which both (a) disconfirms that the
discipline is unique to FEOV, and (b) confirms that the dream/memory store's passive-markdown
persistence is on the weaker side of a distinction the field is actively drawing.

---

## A NEW adversarial finding the lens surfaced (not in the round-0 hypotheses)

**The stationary-evaluator / Goodhart risk applies to sleeper-service AS A LOOP, and the closest
external system names it precisely.** FEOV's red is a FIXED constitution-bound evaluator. RQGM's
central thesis is that self-improvement search "generally assume[s] a stationary evaluation
criterion: a fixed verifier, benchmark, or labeled dataset that remains valid as the agent
improves," and that this "ignores a central feature of evolution." [leaf-verified — arXiv 2606.26294
abstract, accessed 2026-08-23] Sleeper-service is a self-improvement loop whose evaluator (red +
the human's fixed rules) does NOT co-evolve with blue. As `/self-improve` runs daily and blue learns
which arguments pass red, Goodhart pressure accrues: "once a measure becomes a target, systems
optimize specifically for the measure." [search-digest — CACM, "Goodhart's Law Comes for Every
Benchmark You Trust", https://cacm.acm.org/blogcacm/goodharts-law-comes-for-every-benchmark-you-trust/ ,
accessed 2026-08-23] Even the strongest known mitigation (RQGM's held-out human-curated ground
truth) is critiqued as "structurally, a static benchmark at one level higher … Goodhart's Law at a
second order of abstraction." [search-digest — TechTimes coverage of RQGM,
https://www.techtimes.com/articles/319230/20260628/ , accessed 2026-08-23]

FEOV's partial defenses against this are real and worth naming: (a) red carries `memory: project`
of gap patterns, so the evaluator is not fully static — it accumulates new failure classes across
runs (a poor-man's evaluator evolution); (b) the human promotion gate is a held-out check the loop
cannot optimize away; (c) execution proofs anchor to reality, not to red's opinion. But the loop has
NO mechanism to detect that blue has learned to *pass red without getting better*, which is the
exact divergence Goodhart detection needs ("if held-out performance diverges from tuned
performance, you are goodharting"). **File-able:** a periodic held-out audit — a red variant or a
fresh-model reviewer that has NOT seen the accumulating gap-pattern memory — measuring whether
pass-rate improvements track real quality. This is the sharpest borrowable idea the external corpus
offers the in-repo loop, and it is disconfirming by construction.

---

## Lines of inquiry — what I pursued, deferred, and rejected

**Pursued (primary):** Q1 external verification-locus census — the assault above. Outcome: no
single axis novel; differentiator narrows to a COMBINATION; and the self-preference-bias literature
turns partly against FEOV's own escape route.

**Pursued (breadth):** Q4 (gate placement — axis holds, but "gate at promotion" is a converging
pattern, D7); Q5 (execution proof — standard in narrow tasks, the honest claim is the transplant
into open-ended research); Q2 (structural split — outer harness shared, consolidation cores differ
in kind); Q3 (refusable-record — hypothesis holds at the leaf, but bounded to a small write-validator
fix, not a re-architecture).

**Considered and DECLINED this run (recorded so a later run need not re-walk):**
- *AutoGen / CAMEL multi-agent frameworks as a fourth external comparator.* Hypothesis if it paid
  off: they'd show a distinct human-gate or verification posture. Declined: search showed they are
  orchestration substrates (conversation patterns), verification-agnostic — they'd add breadth
  without a new axis, and the multi-agent-debate literature (below) already covers the "do multiple
  agents help?" question more directly.
- *Promptbreeder / Boundless Socratic Learning as self-referential-improvement comparators.*
  Declined for this lane: they bear on the self-improvement axis (Q4) but neither adds a
  verification-locus or gate-placement distinction the six-system census lacks; a later run wanting
  a deeper self-improvement-lineage section should pick these up.

**Weighed and REJECTED as a framing:** "multi-agent debate is FEOV's category, so FEOV inherits
debate's benefits." Rejected on disconfirming evidence: MAD frequently FAILS to beat single-agent
reasoning — sycophancy "collapse[s] debates into premature consensus," homogeneous groups "amplify
shared biases," and one study finds "Isolated Self-Correction Prevails Over Unguided Homogeneous
Multi-Agent Debate" at 2.1–3.4× the token cost. [search-digest — "Peacemaker or Troublemaker: How
Sycophancy Shapes Multi-Agent Debate", arXiv 2509.23055, https://arxiv.org/pdf/2509.23055 ; "The
Cost of Consensus", arXiv 2605.00914, https://arxiv.org/html/2605.00914 ; accessed 2026-08-23]
Rejecting the framing SHARPENS FEOV's positioning: FEOV is explicitly NOT symmetric
consensus-seeking debate — red is chartered to WITHHOLD and blue to ADD, with no "converge to
agreement" step and cross-model seats (fable vs sonnet). The MAD failure literature is thus
evidence FOR FEOV's asymmetric design, and evidence AGAINST describing FEOV as "multi-agent debate."
A synthesizer should position FEOV against MAD's failure modes, not in MAD's lineage.

**Deferred (worth a later run, not this one):** a quantitative proof-of-concept comparing FEOV's
red-PASS-rate trajectory against a held-out reviewer over successive `/self-improve` cycles, to
actually MEASURE the Goodhart divergence the new finding above only argues. That needs run history
this project does not yet have (sleeper-service is Phase-0 scaffold), so it is a real experiment for
a future run, flagged here so the direction is on the record.

---

## Source ledger (for the synthesizer to cite via the tool — access date 2026-08-23 throughout)

Leaf-verified (primary bytes read):
- The Red Queen Gödel Machine, arXiv 2606.26294 — https://arxiv.org/abs/2606.26294
- The AI Scientist, arXiv 2408.06292 — https://arxiv.org/abs/2408.06292
- Gödel Machines (Schmidhuber), IDSIA https://people.idsia.ch/~juergen/gmweb2/gmweb2.html /
  arXiv cs/0309048 https://arxiv.org/pdf/cs/0309048

Search-digest (a synthesizer MUST leaf-verify before citing a load-bearing number):
- Constitutional AI, arXiv 2212.08073 — https://arxiv.org/pdf/2212.08073
- Self-Preference Bias in LLM-as-a-Judge, arXiv 2410.21819 — https://arxiv.org/abs/2410.21819
- Reliability without Validity (LLM-as-Judge), arXiv 2606.19544 — https://arxiv.org/html/2606.19544v1
- Agent-as-a-Judge, arXiv 2410.10934 — https://arxiv.org/pdf/2410.10934
- Reflexion, arXiv 2303.11366 — https://arxiv.org/html/2303.11366
- Darwin Gödel Machine, arXiv 2505.22954 — https://arxiv.org/abs/2505.22954 ; https://sakana.ai/dgm/
- AI Scientist-v2, arXiv 2504.08066 — https://arxiv.org/abs/2504.08066
- AlphaEvolve — https://www.getmaxim.ai/blog/alphaevolve-ai-for-scientific-discovery/
- FunSearch/AlphaEvolve overview — https://richardcsuwandi.github.io/blog/2025/llm-algorithm-discovery/
- Voyager — https://aiunderstanding.org/learn/voyager-and-skill-library-agents
- STORM — https://github.com/stanford-oval/storm
- OpenAI deep research — https://openai.com/index/introducing-deep-research/ ; ByteByteGo https://blog.bytebytego.com/p/how-openai-gemini-and-claude-use
- NeurIPS hallucinated citations — https://fortune.com/2026/01/21/neurips-ai-conferences-research-papers-hallucinations/ ; Nature https://www.nature.com/articles/d41586-026-00969-z
- MLR-Bench, arXiv 2505.19955 — https://arxiv.org/pdf/2505.19955
- MAD sycophancy, arXiv 2509.23055 — https://arxiv.org/pdf/2509.23055 ; Cost of Consensus, arXiv 2605.00914 — https://arxiv.org/html/2605.00914
- safe-self-improvement (OpenClaw) — https://github.com/gateswell/safe-self-improvement-agent
- Self-Improving AI 2026 (eval-gated) — https://www.morphllm.com/self-improving-ai
- Goodhart/benchmark gaming — https://cacm.acm.org/blogcacm/goodharts-law-comes-for-every-benchmark-you-trust/ ; RQGM held-out critique https://www.techtimes.com/articles/319230/20260628/
- Provenance/auditability — Clarus arXiv 2606.30246 https://arxiv.org/pdf/2606.30246 ; Claim-Level Auditability arXiv 2602.13855 https://arxiv.org/html/2602.13855

In-repo design docs (leaf-read at pinned commit 2ce929f):
- plans/memory-architecture.md (§2, §3.1, §6, §7.5, §7.6, §9.4)
- plans/claude-port-plan.md (§3b, §3c)
- plugins/frank-exchange-of-views/skills/research-protocol/SKILL.md
- plugins/prosthetic-conscience/skills/facts-are-fields/SKILL.md
- inputs/run-config.json (model: fable, judgmentModel: sonnet)
