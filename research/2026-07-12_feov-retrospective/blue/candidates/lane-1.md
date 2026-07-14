# Lane 1 — H1 (lane diversity) to saturation, then breadth

Assignment: hypothesis H1 (blue lane diversity is unenforced and lanes converge in practice)
verified deep via direct primary-source comparison of run 2's `blue/candidates/lane-1.md` and
`lane-2.md` against `blue/report.md` and `blue/CHANGELOG.md`; then breadth across H2-H5 plus a
load-bearing discovery outside the frontier (an unmerged fix branch that already resolves much of
H3/H4). 14 web searches/fetches plus extensive local leaf-node verification (git log, git show,
`gh pr view`, grep counts) on this machine; disconfirming budget: 3 of 14 searches spent hunting
evidence that lane convergence is *not* a defect (wisdom-of-crowds independence conditions,
isolated-self-correction-vs-debate literature) — both partially disconfirm the naive "more
diversity is strictly better" reading. Saturation reached on H1: final searches returned
already-seen papers (diversity collapse, artificial hivemind, structural coupling).

**Headline verdict for this lane:** every one of the five frontier hypotheses is **confirmed**,
several more strongly than the frontier itself predicted — but the single most consequential
finding this round is not on the frontier at all: **an open, unmerged pull request
(`ctoforaday/special-circumstances#14`, branch `feat/feov-dogfood-round-1`) already contains
working fixes for the run-1 harness defect, a null-guard for run-2's crash, a zero-token
regression simulator with 11 tests, per-role model routing, a citation ledger, a pre-created
blackboard skeleton (the write-block fix), and a de-DARPA'd Catechism template** — and none of it
has shipped to `main`. Any run invoked against `main` today (including a hypothetical run 3) still
carries R1-HARNESS-1's exact defect class. This reframes "what should change before run 4" from a
research question into, first, a shipping question.

---

## 1. H1 — Lane diversity is unenforced and lanes converge in practice: CONFIRMED, with mechanism

### 1.1 The lane count itself was under-provisioned, not just the differentiation

`commands/research.md` documents `--lanes` default **3**.[^ResearchCommand] Run 2
(`2026-07-12_memory-architecture`) dispatched only **2** lanes — `blue/candidates/` contains
exactly `lane-1.md` and `lane-2.md`, no `lane-3.md`, and `blue/frontier.md`'s own header states
"Lane assignments: lane 1 took H1 to saturation then breadth; lane 2 took H2 to saturation then
breadth" — H3, H4, and H5 were *never independently researched by a dedicated lane at all* in run
2; they were each covered only as "breadth" inside the two hypothesis-deep lanes.[^Run2Frontier]
`workflow.js`'s dispatch loop (`Array.from({ length: lanes }, ...)`) takes `lanes` as a plain
caller argument with no minimum enforced in code.[^WorkflowJs] Whether run 2's `--lanes 2` was a
deliberate operator choice (cost control) or an omission is **not recoverable from the corpus
provided** — there is no run-2 trajectory journal in the retrospective inputs, only run 1's — so
this sub-finding is flagged **unverified-as-to-intent, verified-as-to-fact**: 2 lanes ran where 3
is documented default, and no lane took H3, H4, or H5 as its primary assignment.

### 1.2 Direct comparison, lane-1.md vs lane-2.md: real per-hypothesis depth, real breadth-phase convergence

Read against each other line-by-line, the two lanes are **not** duplicates — each contains
material the other lacks in its assigned-hypothesis-deep section:

- Lane 1 alone: the Open Knowledge Format spec's reserved-file/no-frontmatter rule, the four
  distinct native Claude Code surfaces (`@`-import hop limit, `MEMORY.md` load window,
  `autoMemoryDirectory`, `.claude/rules/`) inventoried individually with citations, the headless-
  hooks open-issue set (#20063/#38651/#40506), the local JSONL transcript schema inspected at the
  leaf node on-machine, and the bidirectional-write-collision analysis of the `memory:` agent
  frontmatter row.
- Lane 2 alone: the two-lever consolidation-corruption fix (candidate-retrieval unspecified;
  expand-invites-rewrite), the LLM-judge-dedup threshold curve (cosine 0.95 vs 0.85-0.87), the
  git-diff-forensic-not-preventive argument with Dependabot/PR-review figures, the
  Auto-Dream-two-writer-conflict analysis, and the RecMem eager-vs-recurrence consolidation
  figures.

Both lanes independently reached, in their **unassigned breadth phase**, the same
top-line findings: (a) memory poisoning / CVE-2026-21852 as a blocking gap absent from the
proposal's §9 — headlined near-identically ("absent from §9 entirely" in both drafts, word-for-
word in lane 1's section title and lane 2's section title); (b) drop-the-confidence-float; (c)
git-diff review as weak/forensic; (d) the same four external alternatives (claude-mem, basic-
memory, mem0/Letta/Zep) with near-identical "steal lists"; (e) headless/`-p` execution risk as an
open concern; (f) the transcript-schema-is-unstable caveat. This is the frontier's predicted
pattern exactly: **assigned-hypothesis material diverges; breadth-phase material converges** —
because "then breadth" gives both lanes the same four remaining hypotheses with no differentiated
method, lens, or source-class assigned to tell them apart once the deep dive ends.[^Lane1][^Lane2]

### 1.3 Synthesis erases the convergence signal entirely — measured, not inferred

`blue/report.md` (2,145 lines, after 4 rounds) contains **zero** occurrences of the strings
"lane-1", "lane-2", "lane 1", or "lane 2" (grep count on this machine, 2026-07-12/13).[^LocalGrep]
The CHANGELOG's Round-0 entry describes the merge in dedup-and-reconciliation vocabulary — "kept
both," "union of," "merged X §n + Y §n," "preserved distinctly-sourced near-duplicates" — with
**no vocabulary class that distinguishes "both lanes independently reached this" from "only one
lane surfaced this."** The memory-poisoning section, for instance — the single most load-bearing
finding in the whole report, independently discovered by both lanes — is merged into §4 with the
same "union of required changes" language used for entirely single-lane content (e.g., lane-1's
OKF reserved-files caveat, never touched by lane 2). A reader of `blue/report.md` alone cannot
tell that the poisoning finding is doubly-corroborated while the OKF caveat is a minority report.
**H1's corollary and H2 (below) are the same defect, confirmed by the same measurement.**

### 1.4 Disconfirming evidence: convergence is not unambiguously bad, and "more diversity" is not a free lunch

Three findings complicate a simple "add more diversity" prescription:

- **Diversity collapse survives persona assignment.** A 2026 study on structural coupling in
  multi-agent LLM systems found that assigning distinct roles/personas **does not** prevent
  convergence — agents "still converge toward homogeneous outputs despite these assignments,"
  because the coupling operates below the level surface role differentiation can reach; the paper
  recommends structural decoupling (less shared context, not just different instructions),
  isolated generation phases, and real-time diversity metrics over persona engineering
  alone.[^DiversityCollapse] This means H4's proposed fix ("distinct lenses/source-classes per
  lane") is directionally right but should not be oversold as a solved problem — it is a mitigant,
  not a guarantee, and the engine should treat lane-count-agreement as *weak* corroboration signal
  even after the fix ships.
- **Wisdom-of-crowds value requires actual statistical independence, which parallel-but-
  unconstrained lanes don't guarantee.** The crowd-wisdom literature is explicit: convergent
  agreement is only informative when errors are *uncorrelated* — "under independence, diversity is
  large and the collective outperforms its members; under correlation, diversity collapses and the
  collective inherits its members' errors."[^WisdomCrowds] Lanes sharing the same base model,
  training data, and (per current design) the same frontier-hypothesis framing are exactly the
  correlated-error case the literature warns about — which sharpens rather than weakens the case
  for provenance tagging (§2 below): un-tagged agreement in this architecture is *presumptively*
  correlated, not independent, until the lanes are engineered to search different source classes.
- **Independent generation beats unguided convergent debate on the numbers that matter, which
  validates FEOV's existing shape more than it undermines it.** A controlled study comparing
  isolated self-correction to homogeneous multi-agent debate found debate induces sycophantic
  conformity (up to 85.5% modal adoption) and "consensus collapse" (oracle-answer gaps up to 32.3
  points), while isolated parallel reasoning matched or beat debate at 2.1-3.4x fewer
  tokens.[^IsolatedCorrection] FEOV's blue-lane phase is already isolated-parallel-then-synthesize,
  not agents-talking-to-each-other-in-real-time — the risky "debate" step in the literature's sense
  is red's cross-round audit-and-rebuttal loop, which is adversarial rather than consensus-seeking
  by design (red's incentive is to find gaps, not agree). This is evidence the overall
  architecture's separation of concerns (isolated build / adversarial gate / judged termination)
  is closer to what the literature recommends than a naive multi-agent debate would be — the
  convergence problem is local to the blue-lane breadth phase, not systemic to the engine.

**H1 verdict: confirmed.** Lane count was under-provisioned this run (2 of a documented default
3); assigned-hypothesis-deep material genuinely diverges; unassigned-breadth material converges
almost completely; the convergence signal is entirely destroyed at synthesis (0 lane-provenance
mentions in blue's report, 0 lane-agreement inputs to red's corroboration grading); and the
literature both validates engineering lane diversity further *and* warns that persona/hypothesis-
order assignment alone (today's only differentiator) is a weak lever compared to structural
decoupling (distinct source classes, distinct search methods, or genuinely isolated context).

---

## 2. H2 — Consensus vs. minority provenance is destroyed at synthesis, not merely under-surfaced: CONFIRMED

This is the natural extension of §1.3/1.4 and the doubts-file's own framing (`doubts.md`: "the
union merge erases claim provenance ... red should weigh them differently"). Verified two ways:

1. **Blue's synthesis vocabulary carries no provenance-count distinction** (§1.3): "kept both" /
   "union of" / "merged" apply identically whether a claim came from one lane or two.
2. **Red's corroboration grading has zero lane-provenance input.** `red/findings.md` (695 lines,
   30+ graded gaps across 4 rounds) contains 66 occurrences of "corroborat-" — all of them grading
   *external citation* confidence (high/medium/low against a web source) — and **zero** occurrences
   of "both lanes," "one lane," "single lane," "lane-sourced," or "lane provenance" (grep count on
   this machine).[^LocalGrepRed] Red's own envelope schema has a `corroboration` array keyed on
   `claim`/`reference`/`confidence` — there is no `lane_count` or `provenance_class` field, so even
   if red wanted to grade lane-agreement, the schema gives it nowhere to put that judgment.

**Consequence, stated precisely for the pragmatist test (likelihood x impact x complexity):** this
is a real, measured design gap (likelihood: certain — it already happened; impact: medium — the
memory-poisoning finding survived on its substantive strength regardless of provenance tagging,
so no wrong conclusion was reached *this run*, but a future run where a single-lane hallucination
and a double-lane finding read identically in prose is a latent miscalibration risk for red and for
the human reader). The backlog's own **CLAIM MANIFEST** proposal is the right-shaped, low-
complexity fix: blue emits a machine-readable ledger (claim -> citation -> self-graded confidence
-> lane provenance) that a synthesizer can consult and a corroboration-grader can weight — "one
artifact, five wins" including this exact doubt, already recognized as approved-direction in
`ideas/backlog.md`, not yet built.[^LiveBacklog] **Grade: High likelihood (the gap is proven, not
speculative) x Medium impact (no verdict was wrong this run, but the miscalibration risk compounds
across runs and topics where corroboration matters more) x Low complexity (an additive field on an
existing schema plus one new file, not new machinery). Recommend: build before run 4.**

---

## 3. H3 — The defect population is bimodal (harness-testable vs. leaf-node-only): CONFIRMED, and largely already solved on an unmerged branch

### 3.1 Every run-1 defect is reproduced today by a working zero-token simulator

Run 1's entire failure (16 agents, 252.9k tokens, 11m48s, terminating UNVERIFIED by honest
deadlock)[^Run1Friction] traces to one caller-side bug: `args` arrived JSON-stringified on a resume
and destructured to `undefined`, so every downstream path was the literal string
`"undefined/..."`.[^Run1Journal] This is **exactly** the class of bug the doubt file predicted would
be "reproducible against a Node simulator that stubs `agent()`." It has, in fact, already been
fixed and tested — not on `main`, but on the open branch `feat/feov-dogfood-round-1` (PR #14,
opened 2026-07-12, still OPEN as of this research, +2281/-46 lines).[^PR14] That branch's
`debate.js` (renamed from `workflow.js` per the backlog's own naming-by-function item) adds, at the
top of the script body:

```
const a = typeof args === 'string' ? JSON.parse(args) : args
const { topic, runDir, lanes = 3, maxRounds = 12, model = null, judgmentModel = null } = a
if (!topic || !runDir || String(runDir).includes('undefined') || String(topic) === 'undefined') {
  throw new Error(`debate: refusing dispatch — topic/runDir unbound (...)`)
}
```

and a companion `tests/simulator/harness.mjs` + `tests/simulator/debate.test.mjs` — a Node
`--test` suite of **11 tests running in ~200ms for zero tokens**, wired into CI as job
`debate-sim`.[^PR14][^SimulatorTests] The harness reproduces the Workflow runtime's `agent()`
contract exactly (label-routing responder, `parallel`/`pipeline` semantics where a throwing thunk
resolves to null rather than rejecting the batch, canned schema-shaped envelopes) and drives the
real script body via `new AsyncFunction(...)` — not a reimplementation, the actual file under test.

**This is independent, primary-source confirmation that H3's proposed mechanism works in practice,
not merely in principle** — it is not lane 1's proposal, it is an artifact that already exists,
already passes, and already sits unshipped.

### 3.2 The existing founding-regression suite, evaluated against what it should contain

The task asks what the founding suite should contain "beyond stringified-args and null-agent-
returns." Reading `debate.test.mjs` in full, it already contains substantially more than those two:

| # | Test | Class |
|---|---|---|
| 1 | Stringified args parse; no `undefined/` leaks into any prompt | Founding (run 1) |
| 2 | Unbound topic/runDir refuses dispatch **before any agent spawns** | Founding (run 1), zero-cost fail-fast |
| 3 | Null `red-merge` return aborts cleanly (not a `TypeError`) | Founding (run 2 crash) |
| 4 | Null `blue-synthesize` return aborts cleanly | Founding (run 2 crash, symmetric case) |
| 5 | Happy path: PASS round 1 -> VERIFIED; phase order; lane count honored | Control-flow baseline |
| 6 | Per-role model routing: bulk seats get `model`, judgment seats inherit unless `judgmentModel` set | New efficiency lever (run-3 doctrine) |
| 7 | Contested docket: a re-raised gap reaches the judge exactly once; un-recurred gaps stay off the docket; adjudicated gaps leave red's verdict scope | Core termination logic |
| 8 | Judge `deadlock: true` ends the debate UNVERIFIED with the deadlock stamp | Termination path (this is what actually fired in run 1!) |
| 9 | Safety ceiling: always-new gap ids never trigger the judge; ceiling stamps assembly at `maxRounds` | Termination path |
| 10 | Citation passes scale 1..4 with `claim_count`, and carry the ledger clause | Cost-scaling logic |
| 11 | Friction aggregates from every seat with correct attribution string | Self-improvement input integrity |

Measured against the task's own bar ("beyond stringified-args and null-agent-returns"), **this
suite already clears that bar** — it covers exactly the control-flow surface doubts.md and the
frontier worried about (deadlock, contested docket, safety ceiling) plus two newer levers
(per-role models, citation ledger) the corpus given to this retrospective does not even mention,
because the retrospective's `backlog.md` input snapshot **predates** them (see §4.2).

**What it does NOT and structurally cannot cover** (correctly left out, per H3's own bimodal
prediction):

- The **write-block false positive** on `report.md`/`findings.md` filenames — this is a
  *tool-permission* decision inside the execution environment, not a branch in `debate.js`'s
  control flow; the simulator's `agent()` stub cannot exercise a real Write-tool guard. Confirmed
  fix path is different in kind: PR #14 sidesteps it by having `/research` **pre-create the
  blackboard skeleton** (stub headers for `blue/report.md`, `red/findings.md`, etc.) so subagents
  only ever `Edit`/append, never `Write` a fresh file with a trigger-word name — this is a
  live-smoke-testable fix (verify the guard doesn't fire against the pre-created files under real
  tool permissions), not a unit-testable one.[^PR14Description]
- **ENAMETOOLONG on Windows heredocs** — an OS/shell command-length limit hit by real Bash
  invocations; unit-testable only in the trivial sense of asserting a chunking helper's math, not
  in the sense of reproducing the actual OS error. **Live-smoke-testable at best** (a smoke run
  that deliberately writes a >8KB file via heredoc on Windows CI).
- **PDF-table/full-text extraction, primary-security-advisory access, live-citation drift (star
  counts, issue statuses)** — these are leaf-node fidelity problems against real external sources;
  by construction **only observable in production** (or against real network state in a live-smoke
  run), because the simulator's entire value proposition is *not calling real tools*.
- **Auto Dream runtime behavior** (server-side flag) and the **Springer auth-wall** — both require
  live external state the simulator cannot stub meaningfully (stubbing "what would the real
  feature output" begs the question the audit exists to answer).

This exactly reproduces H3's predicted bimodal split, with one refinement the frontier didn't
anticipate: there is a **third class in the middle** — write-block and ENAMETOOLONG are neither
zero-token-unit-testable nor "only observable in production"; they are **live-smoke-testable**:
reproducible cheaply with 1 lane / 1 round / a cheap model against the real tool permissions and
real OS, without needing a full multi-round live run. The task's own three-way classification
(unit / live-smoke / production-only) is therefore the right taxonomy, and the corpus already
sorts cleanly into it once you have PR #14's evidence in hand:

- **Zero-token unit-testable (simulator):** args parsing/guard, null-agent-return handling, round
  loop, contested docket, deadlock, safety ceiling, citation-pass-count scaling, per-role model
  routing, friction attribution, lane-count dispatch. (All 11 current tests, plus recommend adding:
  a lane-count-floor assertion per §1.1, and a claim-manifest-provenance-passthrough test once §2's
  fix ships.)
- **Live-smoke-testable (1 lane, 1 round, haiku, real tools/OS):** write-block false positives
  against real filenames, ENAMETOOLONG-class heredoc limits, hook-fire-under-headless behavior
  (flagged as a proposal-specific but methodologically identical concern in the memory-architecture
  corpus itself[^Lane1]), the pre-created-blackboard-skeleton fix's actual effect on Write-tool
  behavior.
- **Only observable in production (full multi-round live run, or genuinely unavailable without new
  tooling):** PDF-table/full-text extraction fidelity, primary-security-advisory reachability,
  live-source drift (star counts, issue statuses, package pivots) between citation and
  verification, Auto-Dream-class server-flag-gated feature behavior, paywalled-primary-source
  access (Springer/auth-wall class).

**Grade for "adopt PR #14's simulator design as the model going forward": High likelihood (already
built and passing) x High impact (would have caught run 1's entire 252.9k-token failure before any
agent spawned, per the test itself) x Low complexity (it exists; the only remaining cost is
review and merge). This is not a proposal — it is a shipping decision.**

---

## 4. H4 — The highest-leverage pre-run-4 changes are structural and cheap: CONFIRMED, and a significant fraction is already built

### 4.1 What the run-2 corpus alone would have recommended

Reasoning from the retrospective's `inputs/` corpus in isolation (as if PR #14 did not exist), the
cheap-structural fixes are: (a) a caller-side arg-shape preflight guard — confirmed already fixed
run-1-side by PR #14 (§3.1); (b) an artifact-filename allowlist/carve-out for the write-block
heuristic — confirmed already addressed by PR #14's pre-created-skeleton approach (§3.2), a
different but equally cheap mechanism than the allowlist the doubt file speculated about; (c)
engineered per-lane diversity (distinct lenses/source-classes) — **not yet built anywhere**, this
remains open (§1.4 grades it Medium-effort given the diversity-collapse literature's caution that
persona-only fixes under-deliver).

### 4.2 What the live repository state adds, that the retrospective's input snapshot does not mention

The `inputs/backlog.md` given to this retrospective is a **stale snapshot**: comparing it against
`ideas/backlog.md` on `main` at HEAD (`git show 9ff0fad`), four items marked open in the
retrospective's copy are marked `[x]` DONE on `main` as of commit `9ff0fad`
("docs(backlog): graduate simulator, per-role models, citation ledger, write-block fix to PR
#14"):[^LiveBacklog]

- The zero-token workflow simulator (§3.1-3.2) — DONE on PR #14 branch, not yet merged.
- The `agent()` null-guard (run 2's actual crash cause: `TypeError` on `redEnv.verdict` when a
  quota wall killed red mid-round) — DONE on PR #14 branch.
- Per-role model routing (`model` for bulk seats, `judgmentModel` for judgment seats, defaulting to
  inherit-session) — DONE on PR #14 branch; per the code comment this reflects an explicit
  efficiency doctrine ("cheapen redundancy and mechanics, never judgment or the adversary").
- The write-block fix via pre-created blackboard skeleton — DONE on PR #14 branch (first live
  trial pending — the branch's own commit message says "first live trial is run 3").
- The citation ledger (`red/citation-ledger.md`: high-confidence verifications persist across
  rounds unless the CHANGELOG shows the section changed) — DONE on PR #14 branch, addressing the
  backlog's own observation that "red's full-report re-read x scaled lens instances x round is the
  dominant burn."
- The Heilmeier-is-DARPA-shaped doubt (doubts.md item 2) — **already independently identified and
  partially fixed**: PR #14 replaces `heilmeier_template.md` with `catechism_template.md`, keeping
  questions 1-3 (what/how-today/what's-new) but reframing questions 4-9 into topic-agnostic
  worth-our-time framing: "The case against" (every honest objection at full strength), "Of
  interest, or merely interesting?", "What changes if it works — and what happens if we simply
  don't do it?", and cost/stopping-points.[^CatechismTemplate] This reads naturally for an
  architecture-evaluation topic (verified by inspection: none of the reframed questions presuppose
  a funding ask, a program, or a deliverable date) in a way the original Heilmeier's "how long will
  it take?" / "who cares?" phrasing does not.

**None of this has reached `main`.** PR #14 is open, unmerged, dated 2026-07-12, same day as run
1's defect. The retrospective's own input corpus (frozen presumably at task-setup time) does not
reflect it. **This is the single highest-leverage "change before run 4" available, and it costs
review time, not engineering time — the work is done.**

### 4.3 What remains genuinely open (not fixed by PR #14), graded

| # | Proposal | Likelihood | Impact | Complexity | Grade / recommendation |
|---|---|---|---|---|---|
| 1 | Merge PR #14 | n/a (already built) | High (fixes 2 of run 1-2's root-cause defects; unlocks the simulator, ledger, per-role costs, write-block fix) | Low (review only) | **Do this before anything else in this list** |
| 2 | Engineered per-lane diversity: assign distinct source-classes or search methods, not just hypothesis order, and do not rely on persona/role text alone | High (run 2 measured the convergence directly) | Medium (breadth-phase convergence didn't produce a wrong verdict this run, but the corpus's own literature warns persona-only fixes under-deliver) | Medium (requires redesigning the lane-dispatch prompt template, per-lane; not a script change, a protocol change) | Fix, but scope it to source-class assignment (e.g., lane N searches vendor blogs + issue trackers, lane N+1 searches academic literature + competitor products) rather than persona theater, per §1.4's disconfirming evidence |
| 3 | Claim manifest with lane provenance (H2 fix) | High (gap is proven) | Medium | Low (additive schema field + one file) | Fix before run 4 — cheapest fix on this list relative to the risk it retires |
| 4 | Lane-count floor: assert `lanes >= 3` (or require an explicit justified override) at dispatch | Medium (only matters if under-provisioning recurs; §1.1's intent is unrecoverable from this corpus) | Medium (an under-provisioned run silently loses an entire hypothesis's dedicated attention) | Low (one guard clause, same shape as the args guard already on PR #14) | Fix — cheap enough that even uncertain intent doesn't justify skipping it |
| 5 | Append-friendly write path / larger Bash spawn budget for ENAMETOOLONG | Medium (Windows-specific, recurred once) | Low-Medium (forced ~6 append calls, not a failure, just overhead) | Medium (a new tool surface, not a prompt change) | **Risk-accept for now**: the workaround (heredoc chunking) already works; building a new write-append tool to save a few extra tool calls is exactly the "design made strictly worse to satisfy an edge case" the pragmatist test warns against, *unless* the pre-created-blackboard-skeleton fix (append-not-write) already moots most of this by construction — recommend re-measuring after PR #14's live trial before spending anything here |
| 6 | Round-scoped audit narrowing (round 1 full leaf-node; rounds 2+ changed-sections-only + contested gaps + spot checks) | Medium (would cut red's dominant cost driver) | Medium-High (cost) | Medium (trades against the full-re-read principle explicitly; the backlog itself marks this "human-gated," not yet approved) | **Hold** — correctly already flagged in the live backlog as needing an explicit human call because it trades against a stated design principle; not this lane's call to make unilaterally |

---

## 5. H5 — The friction corpus is dominated by 2-3 systemic gaps, not a long tail: CONFIRMED, ranked by distinct-role count

Counting distinct agent-role attributions across `run1-friction.md` and `run2-friction.md`
(role = red-auditor / blue-researcher / lead-judge; a role counts once per complaint class
regardless of how many rounds within that class it repeats in):

| Rank | Gap | Distinct roles reporting | Rounds/runs span | Status |
|---|---|---|---|---|
| 1 | **PDF full-text / table-extraction** (arXiv/HTML fetches lossy for in-table numbers) | red, blue, judge (3/3) | Run 2, rounds 1-4 (every round) | Open — highest-value unbuilt tool |
| 2 | **Primary security-advisory access** (CVE-2026-21852 remediation detail, post-cutoff vendor-blog-only sourcing) | red, blue, judge (3/3) | Run 2, rounds 2-3 | Open |
| 3 (tie) | **Uninitialized run-directory/topic variables** ("undefined" paths) | red, blue, judge (3/3) | Run 1, all rounds | **Fixed on PR #14 (unmerged)** |
| 3 (tie) | **Report-named-file write-block** (report.md/findings.md heuristic) | red, blue, judge (3/3, blue in run 1, red in run 2, judge referencing both) | Run 1 + Run 2 round 1 | **Fixed on PR #14 (unmerged)** |
| 5 | Windows ENAMETOOLONG on heredoc writes | red (1/3) | Run 2, round 1 | Open, workaround exists (risk-accept per §4.3) |
| 6 | Live-source drift (star counts, issue statuses, package-behavior pivots) needing recorded access-date deltas | red (1/3) | Run 2, round 1 | Open, protocol-documentation fix, not a tool gap |
| 7 | No way to inspect the (unbuilt) trajectory-extractor implementation | red (1/3) | Run 2, round 3 | Open, but the underlying artifact doesn't exist yet either |
| 8 | No sandbox to observe Auto Dream's actual runtime behavior (server-flag-gated) | red (1/3) | Run 2, round 4 | Open, external dependency, not fixable by this repo |
| 9 | Springer/institutional-access paywall (auth-wall) blocking one source | red (1/3) | Run 2, round 4 | Open, single-occurrence |

**H5 verdict: confirmed, with a sharper picture than the frontier predicted.** The frontier
guessed "2-3 systemic gaps ... while the write-block/ENAMETOOLONG/preflight-guard complaints
cluster in round 1 only (already fixed)." The corpus shows **four** items at the top tier by
distinct-role count (not 2-3), but two of those four (the harness-plumbing pair) are correctly
identified as already-addressable-and-now-addressed (on the unmerged branch), leaving exactly the
frontier's predicted **stable, unresolved short list of two**: PDF/table extraction and primary-
source (security-advisory) access, both dominated by document-fetch fidelity rather than tool
diversity, exactly as H5 predicted. The backlog's own text independently reaches the identical
ranking ("TOP TOOL GAP, requested by red, blue, AND judge across all 4 rounds:
PDF full-text/table extraction").[^LiveBacklog]

---

## 6. Consolidated proposal list for this lane (feeds synthesis)

| # | Proposal | Likelihood | Impact | Complexity | Disposition |
|---|---|---|---|---|---|
| 1 | Merge PR #14 (args guard, null guards, simulator, per-role models, citation ledger, write-block fix, Catechism template) | n/a | High | Low (review only) | **Do first** |
| 2 | Claim manifest with lane-provenance tagging (consensus vs. minority) | High | Medium | Low | Fix before run 4 |
| 3 | Engineered per-lane diversity by source-class/method, not persona text | High | Medium | Medium | Fix, scoped narrowly per §1.4 |
| 4 | Lane-count floor / explicit-override guard | Medium | Medium | Low | Fix |
| 5 | Append-friendly write path for ENAMETOOLONG | Medium | Low-Medium | Medium | Risk-accept pending PR #14's live trial |
| 6 | Round-scoped audit narrowing (rounds 2+) | Medium | Medium-High | Medium | Hold — needs explicit human call (trades against full-re-read principle) |
| 7 | PDF full-text/table-extraction tool | High (recurs every round) | High (blocks definitive verdicts on multiple cited figures) | Medium-High (new tool or MCP) | Fix — highest-ranked friction item |
| 8 | Primary security-advisory fetch/access path | Medium (narrower applicability than #7) | High when it recurs (load-bearing for CVE-dependent claims) | Unclear (may require an allowlisted domain or an authenticated fetch capability outside this repo's control) | Fix if feasible; else document as a standing risk-accept with the corpus's own hedging pattern (§4 of run 2's report) as the template |

---

## Footnotes

[^ResearchCommand]: *Run a frank exchange of views* — `commands/research.md`, frank-exchange-of-views plugin, this repo, accessed 2026-07-13. "`--lanes` (blue candidate drafts, default 3)".
[^Run2Frontier]: `research/2026-07-12_memory-architecture/blue/frontier.md`, this repo, accessed 2026-07-13. "Lane assignments: lane 1 took H1 to saturation then breadth; lane 2 took H2 to saturation then breadth."
[^WorkflowJs]: `plugins/frank-exchange-of-views/skills/research-protocol/scripts/workflow.js`, this repo (branch `port-plan-review`, current `main`-equivalent), accessed 2026-07-13. `const { topic, runDir, lanes = 3, maxRounds = 12 } = args`; `await parallel(Array.from({ length: lanes }, ...))` — no minimum-lanes assertion.
[^Lane1]: `research/2026-07-12_memory-architecture/blue/candidates/lane-1.md`, this repo, accessed 2026-07-13.
[^Lane2]: `research/2026-07-12_memory-architecture/blue/candidates/lane-2.md`, this repo, accessed 2026-07-13.
[^LocalGrep]: Local verification, this machine, 2026-07-13: `Grep "lane-1|lane-2|lane 1|lane 2"` against `research/2026-07-12_memory-architecture/blue/report.md` (2,145 lines) → 0 matches.
[^LocalGrepRed]: Local verification, this machine, 2026-07-13: `Grep "both lanes|one lane|single lane|lane-sourced|lane provenance"` against `research/2026-07-12_memory-architecture/red/findings.md` (695 lines) → 0 matches; `Grep "corroborat"` on the same file → 66 matches, all against external-citation confidence, none against lane count.
[^DiversityCollapse]: *Diversity Collapse in Multi-Agent LLM Systems: Structural Coupling and Collective Failure in Open-Ended Idea Generation*, arXiv:2604.18005, accessed 2026-07-13. Role/persona assignment does not prevent convergence; recommends structural decoupling, isolated generation phases, real-time diversity metrics.
[^WisdomCrowds]: Search synthesis across *The Wisdom of the LLM Crowd* (alexanderakm.github.io) and related 2026 wisdom-of-crowds/LLM-ensemble literature, accessed 2026-07-13. "Under independence, diversity is large and the collective outperforms its members; under correlation, diversity collapses and the collective inherits its members' errors."
[^IsolatedCorrection]: *The Cost of Consensus: Isolated Self-Correction Prevails Over Unguided Homogeneous Multi-Agent Debate*, arXiv:2605.00914, accessed 2026-07-13. Isolated self-correction matches/beats debate at 2.1-3.4x fewer tokens across three 7-8B models; debate induces sycophantic conformity (up to 85.5%) and consensus collapse (oracle gaps up to 32.3 points).
[^Run1Friction]: `research/2026-07-12_feov-retrospective/inputs/run1-friction.md`, this repo, accessed 2026-07-13. "Cost of the null run: 16 agents, 252.9k tokens, 11m48s."
[^Run1Journal]: `research/2026-07-12_feov-retrospective/inputs/run1-defect-record/trajectories/journal.jsonl`, this repo, accessed 2026-07-13. Raw agent transcripts showing every dispatch receiving literal `undefined` topic/runDir paths, and each agent's refusal to fabricate research against them.
[^PR14]: `gh pr view 14` (ctoforaday/special-circumstances), accessed 2026-07-13. Title "feat: template-misfit friction + dogfood round-1 fixes", state OPEN, +2281/-46, branch `feat/feov-dogfood-round-1`.
[^SimulatorTests]: `git show feat/feov-dogfood-round-1:plugins/frank-exchange-of-views/tests/simulator/debate.test.mjs` and `.../harness.mjs`, this repo, accessed 2026-07-13. 11 `node --test` cases; `AsyncFunction`-wrapped script-body execution against a stubbed `agent()`/`parallel`/`pipeline` world.
[^PR14Description]: `gh pr view 14` body text, accessed 2026-07-13. "`/research` now pre-creates the blackboard skeleton so subagents only append (dodges the harness write-block on fresh report-like files; red's own recommended fix)."
[^LiveBacklog]: `git show 9ff0fad:ideas/backlog.md` (main @ HEAD, this repo), accessed 2026-07-13, diffed against `research/2026-07-12_feov-retrospective/inputs/backlog.md` (the retrospective's frozen snapshot) — four items flip from open to `[x]` DONE (simulator, null-guard, per-role models, write-block fix) between the snapshot and HEAD; commit message "docs(backlog): graduate simulator, per-role models, citation ledger, write-block fix to PR #14".
[^CatechismTemplate]: `git show feat/feov-dogfood-round-1:plugins/frank-exchange-of-views/skills/research-protocol/references/catechism_template.md`, this repo, accessed 2026-07-13, compared against `plugins/frank-exchange-of-views/skills/research-protocol/references/heilmeier_template.md` (current `main`). Catechism reframes questions 4-9 into "the case against," "of interest vs. merely interesting," "both sides of the ledger," and "cost and stopping points" — topic-agnostic, unlike the original's "who cares?"/"how long will it take?" funding-pitch framing.
