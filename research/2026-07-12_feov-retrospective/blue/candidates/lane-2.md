# Blue lane 2 — H2 (consensus-vs-minority provenance) deep, then breadth

Scope: H2 (does the union-merge preserve which claims were 1-lane vs 2-lane sourced, and does
red's grading use lane-count-agreement as an input?) to saturation, then breadth across H1, H3,
H4, H5. Primary-sourced against the run-2 corpus (`2026-07-12_memory-architecture/`: both lane
candidates, `blue/CHANGELOG.md`, `blue/report.md`, `red/findings.md`, `debate.md`, `report.md`),
the engine's own source (`workflow.js`, `agents/blue-researcher.md`, `agents/red-auditor.md`),
the run-1 harness-defect trajectory (`journal.jsonl`, `debate.md`), the repo's live
`ideas/backlog.md`, and 4 web searches (disconfirming budget: 2 of 4, on multi-agent-debate
complexity-cost literature and on PDF-extraction tool availability — see §6, §7).

---

## 1. H2 — CONFIRMED: consensus-vs-minority provenance is destroyed at synthesis, not merely under-surfaced

### 1.1 The merge vocabulary carries no per-claim provenance tag

Read against `blue/CHANGELOG.md`'s Round-0 entry: every merge operation is described in
class-level language — "deduplicated overlapping claims," "merged risk tables," "preserved
distinctly-sourced near-duplicates" — but the language groups *sections*, never tags individual
*claims* with which lane(s) reached them.[^ChangelogR0] The synthesized `blue/report.md` body
mentions "lane" exactly 5 times in 2972 lines, all at the method-statement or table-header level
("two research lanes," "consolidated, both lanes"), never as an inline per-claim
attribution.[^BlueReportGrep] The one exception is negative provenance — §10 flags items
"cited by the proposal without independent corroboration in either lane" — which shows the
synthesizer *can* express lane-count when it chooses to, but does so only for the
zero-lane case, never for the one-lane (minority) case.[^BlueReportUnverified]

### 1.2 Red's corroboration grading has no lane-count-agreement input — confirmed by direct search

`red/findings.md` (695 lines, cumulative across 4 rounds) mentions "lane" exactly once, in a
housekeeping line: "Disconfirming budget met in both blue lanes. Not a gap."[^RedFindingsGrep]
No corroboration-confidence grade anywhere in the file cites "found independently by both lanes"
or "single-lane-sourced" as a factor. Red's own stated grading dimensions
(`agents/red-auditor.md`: "for each statement↔reference pair YOU MUST assign a corroboration
confidence") are entirely about the *external* reference, never about *internal* lane
provenance.[^RedAuditorSpec] This is a clean confirmation of H2's corollary: red is regrading
every claim blind to whether it was 1-lane or 2-lane sourced, exactly as predicted.

There is one adjacent, easily-conflated finding that is *not* the same gap and should not be
credited to H2: red's R2-10 caught blue's own cross-lane-comparison being laundered as *external*
corroboration — footnote `[^SingleUserLowRisk]` cited "practitioner consensus surveyed
2026-07-13" for a claim that was actually blue's own reasoned synthesis, not an outside
source.[^R2-10] Red fixed a citation-provenance failure (internal reasoning presented as
external fact), not a lane-provenance failure (claim origin invisible to the reader). The two are
easy to blur — both are "where did this claim really come from" questions — but the fix for one
(attribute-or-relabel a footnote) does not fix the other (tag a claim's lane origin). Distinguishing
them matters because a fix aimed at R2-10's failure mode (tighter footnote discipline) would not
have caught H2's failure mode at all, and a reader who assumes red's citation lens already covers
"claim provenance" will wrongly conclude H2 is handled.

### 1.3 Quantifying what convergent vs. lane-unique content actually looked like (feeds H1, §2)

Comparing `blue/candidates/lane-1.md` (355 lines, H1-deep) against `lane-2.md` (321 lines,
H2-deep) directly:[^Lane1][^Lane2]

**Both lanes independently reached** (true consensus — strong signal): CVE-2026-21852 memory
poisoning as the "absent from §9" blocking omission (both lanes headline this identically as a
new, un-enumerated risk); the core alternatives roster (claude-mem, basic-memory, mem0, Letta,
Zep) with near-identical dispositions; OKF spec verification (v0.1 Draft, `type` the only
required field); the transcript JSONL substrate at the leaf node; `@`-import semantics (4-hop
max, silent-disable on declined approval); `.claude/rules/` as an unconsidered projection
alternative; the confidence-float-removal recommendation; Letta sleep-time compute and Stanford
generative-agents importance-threshold citations for H3/cadence.

**Lane-1-only** (minority — real, not noise): local leaf-node repo verification catching two
*false premises* in the proposal (the secret-scrub gate does not exist; `sleeper-service`'s
`docs/scheduling.md` does not exist) — a critical-stance move only this lane performed; the open
GitHub issues on hooks misbehaving under `claude -p` (#20063, #38651, #40506); the Dependabot
54%-merge-rate bot-review-fatigue citation; the `autoMemoryDirectory`-into-store ingest-collapse
idea.

**Lane-2-only** (minority — real, not noise): RecMem's eager-vs-recurrence consolidation finding
(77–87% wasted tokens, no accuracy gain from eagerness); the specific dedup-recall mechanics
(paraphrase-detection gap, LLM-judge sensitivity threshold at cosine 0.85–0.95); native "Auto
Dream" as a rolling-out competing feature (H5 revised into a scope-collision finding); the
headless parallel-Task-fan-out hang under non-TTY parents (#56540); the BeliefMemory/ALFWorld
confidence-helps counter-evidence.

Every one of these lane-unique findings survived into the final `blue/report.md` untouched by
provenance loss in the sense that the *content* made it through — union-not-summary held for
substance. What was lost is the *metadata*: a reader of `report.md` cannot tell, without doing
the archaeology above, that the false-premises catch and the RecMem citation are each
one-agent-hours of independent leaf-node work (minority reports, high-value, unreplicated) versus
the CVE finding being independently triangulated by two agents running different search strategies
(consensus, cross-validated). Red's corroboration-confidence field is the natural home for this
signal and currently has no column for it.

### 1.4 The fix already exists as a proposal — verify it against the actual failure mode

The repo's own live backlog (not the frozen `inputs/backlog.md` snapshot — the current
`ideas/backlog.md`) independently proposes exactly the missing mechanism, unprompted by this
retrospective: a **claim manifest** — "blue emits a machine-readable ledger (claim → citation →
self-graded confidence → lane provenance)," justified as "consensus-vs-minority doubt answered
via provenance; ... red grading gets provenance input."[^ClaimManifest] This is independent
corroboration that the operator/system already converged on H2's diagnosis before this
retrospective started — the doubt is not merely confirmed by evidence, it was already scheduled
for a fix. **Recommendation for §3 (pre-run-4 changes, H4): promote the claim manifest from
"idea" to a graded proposal** (see §5, item 1) rather than re-deriving a new mechanism; it is the
right fix at the right layer (blue emits it at synthesis time, red consumes it as a grading
input) and should be built once, not proposed twice.

**Complexity check (blue's own pragmatist duty):** a full claim manifest requires blue to track
per-claim lane origin through the union merge — real authoring discipline, not free. The cheaper
partial version — tag only claims present in exactly one candidate draft (minority) at synthesis
time, using a simple set-difference over section content, and leave doubly-sourced claims
untagged (silence = consensus by default) — captures the highest-value half (flagging
single-source claims for extra scrutiny) at a fraction of the bookkeeping cost of a full
citation → confidence → provenance ledger for every claim. Recommend the cheap half first;
promote to the full manifest only if red's round-1 usage shows the minority tag is
load-bearing for a real verdict change.

---

## 2. H1 — breadth: lane diversity is real but *unengineered*, not absent (REFINES the frontier hypothesis)

The frontier's H1 predicted "substantial structural and substantive overlap ... a CHANGELOG
dominated by dedup-and-merge operations rather than reconciliation of genuinely distinct
material." §1.3's direct comparison above shows this is **half-confirmed, half-refuted**:

- **Confirmed:** both lanes converge tightly on the shared "obvious" material — the alternatives
  survey, the CVE, the native-surface documentation facts. This is expected and *not itself a
  diversity failure* — a small, well-documented field (5-ish credible alternatives, one disclosed
  CVE) has a limited number of "obvious" citations, and two competent researchers should find the
  same ones. Convergence on the field's canonical facts is a *good* sign (independent
  replication), not evidence of redundant lanes.
- **Refuted:** the CHANGELOG is not, in fact, dominated by dedup-and-merge with no reconciliation
  of distinct material — §1.3's minority-content lists show each lane surfaced substantial,
  load-bearing, non-overlapping findings (false-premise repo verification vs. RecMem/dedup
  mechanics vs. Auto Dream scope-collision) that the other lane did not find. The "breadth" phase
  did not collapse into redundant re-coverage of the same H1-H5 space; it produced different
  breadth.
- **The lane-count observation stands as a separate, valid finding.** Only 2 lanes were
  dispatched for the memory-architecture run though `commands/research.md`'s documented default
  is 3 (`--lanes` flag, default N=3).[^ResearchCmd] This retrospective run itself was *also*
  dispatched with 2 lanes (this document is lane 2 of 2, not lane 2 of 3) — the same
  under-provisioning the frontier flagged in the subject run recurred in the run studying it.
  Whether 2 or 3 is "enough" cannot be settled by this evidence alone (no 3-lane run exists to
  compare against), but the mechanism that produces diversity — different starting hypotheses,
  not different methods — means each *additional* lane's marginal diversity depends on how many
  hypotheses remain unassigned as "first," which argues for lane count tracking hypothesis count
  (5 hypotheses → 3-5 lanes gets each a dedicated first-assignment; 2 lanes leaves 3 hypotheses
  never someone's deep-dive, covered only as everyone's shallow breadth pass).

**Disconfirming check (multi-agent-debate literature):** diversity gains from additional agents
plateau quickly in the wider multi-agent-debate literature — accuracy improvements from
diversity/committee-size "typically plateau with 2-3 debate rounds and 2-4 agents," with
diminishing or negative returns past that band, and correlated-error reduction requires *genuine*
role/method diversity, not just headcount.[^DiminishingReturns] This cuts against "just add more
lanes": the frontier's own framing ("worth checking whether lane count itself was
under-provisioned") risks scope creep if the fix is "dispatch more lanes" rather than "assign
each lane a genuinely distinct method." The evidence in §1.3 supports the *method*-diversity
reading over the *headcount* reading: lane-1's distinctive contribution (false-premise repo
verification) came from a **local-critical-stance lens** neither lane was explicitly assigned but
lane 1 happened to apply; lane-2's distinctive contribution (dedup mechanics, native-feature
scope collision) came from **treating the assigned hypothesis's own literature more
deeply** (RecMem, LLM-judge-dedup papers a "breadth" pass would likely have skipped). This
suggests the cheap fix is **assigning each lane an explicit method/lens** (e.g., "lane N also
runs a local-repo critical-stance pass: verify every claim the proposal makes about *this*
codebase's current state against the actual files") rather than **assigning more lanes** — same
diversity gain, no added agent-dispatch cost. See §5 item 2.

---

## 3. H3 breadth — defect population is *trimodal*, not strictly bimodal: unit-testable / live-smoke-testable / only-observable-in-production

The frontier's H3 framed the population as bimodal (harness-plumbing bugs vs. leaf-node/citation
bugs). Classifying every defect actually recorded across both runs shows a real **third tier** —
cheap, tool-level checks that need a real environment call but no model reasoning and no full
debate run — that the bimodal framing collapses into one of the other two and shouldn't.

### 3.1 Zero-token unit-testable (Node simulator stubbing `agent()`)

- **Uninitialized/stringified run-directory and topic (R1-HARNESS-1).** Root cause verified as
  caller-side: `const { topic, runDir, lanes = 3, maxRounds = 12 } = args` destructures a
  wrongly-typed `args` (arrived JSON-stringified on a resume) to `undefined`, which then
  interpolates as the literal string `"undefined"` into every downstream prompt.[^Workflow]
  Reproduced in every one of 16 agent dispatches across the run-1 defect trajectory — frontier,
  4 parallel lane attempts, synthesis, 3 red lens attempts, red-merge, judge, and final
  assembly all independently detected and refused to fabricate against it.[^Journal] Purely a
  caller-side value-shape bug; 100% reproducible by feeding the simulator a JSON-stringified
  `args` blob and asserting the constructed prompts contain no `"undefined"` substring.
- **`redEnv` null crash.** `agent()` returns `null` on terminal failure (e.g., a quota wall);
  run 2 crashed with `null is not an object (evaluating 'redEnv.verdict')` — already recorded as
  a fixed, closed backlog item and named as the simulator's "second founding test case."[^Backlog]
- **Round/deadlock/contested-docket control flow.** `workflow.js`'s loop logic — `contested`
  (set-intersection of current and previous gap ids), `hasNew`, the judge dispatch gate
  (`if (contested.length > 0)`), and the deadlock/exhausted/verdict computation at the end — is
  pure control flow over canned envelope shapes and needs zero live agent calls to test.
- **Citation-pass-count arithmetic.** `Math.min(4, Math.max(1, Math.ceil((blueEnv.claim_count ||
  20) / 40)))` is pure arithmetic with untested boundaries (0, missing, 40, 41, 160+) — a
  one-line assertion table.
- **The one round-2 defect not yet in the backlog's founding list:** the round where red-merge
  had **zero round-2 lens-pass inputs by construction**, because the upstream round-2 lens
  dispatches also hit the same undefined-path defect.[^RedMergeR2] The backlog's existing
  founding-case list ("clean PASS, multi-round FAIL→revision, contested docket, judged deadlock,
  safety ceiling, stringified args, unbound args" plus null-returns)[^Backlog] does not name this
  scenario explicitly — it is a *cascading* failure (stage N's precondition silently violated by
  stage N-1's failure) distinct from a single-call null return, and the workflow script has no
  explicit check that `red/candidates/round-N-*` is non-empty before dispatching red-merge. Add
  it as a named regression: "red-merge dispatched against an empty candidates directory should
  [refuse / surface a distinct blocked state], not silently produce a plausible-looking envelope."
- **Gap-id rollover across non-adjacent rounds.** `prevGapIds` is reassigned every round to
  *only* the current round's `gapIds` — so a gap closed in round 1 that resurfaces (same id) in
  round 3 is compared only against round 2's ids, not the full adjudicated history, and would be
  classified "new" rather than "contested" even though it is a recurrence of an already-ruled
  gap. This is exactly the kind of multi-round-only defect that is expensive to trigger live
  (needs 3+ real debate rounds to even present) and cheap to simulate (three canned `redEnv`
  objects with a repeating id) — a second not-yet-named founding case.
- **Edge lane/round counts.** `--lanes 0`, `--lanes 1`, `--maxRounds 0` are all currently
  unguarded; `lanes=0` produces `parallel([])` (no-op) then a synthesis call reading zero
  candidate files, and `maxRounds=0` skips the round loop entirely while producing a verdict
  message ("debate ended: UNVERIFIED after 0 round(s)") indistinguishable from a real 0-round
  deadlock. Recommend the simulator assert on the *log message*, not just the returned envelope,
  since operators reading the transcript rely on that line to disambiguate "never ran" from
  "ran and failed at round 0."
- **Malformed-but-non-null envelope defensive gap (open question, not yet resolved here).**
  `workflow.js` passes a JSON-schema (`RED_ENVELOPE`, `BLUE_ENVELOPE`, `JUDGE_ENVELOPE`) to
  `agent()` but the script itself does no runtime shape validation before destructuring
  (`redEnv.gaps.map(...)` would throw on a `gaps`-less object). Whether the Workflow tool's
  schema parameter *guarantees* conformance-or-null (making this defensive code genuinely
  unnecessary) is a fact about the harness this retrospective cannot verify from inside the
  plugin's own source — flagged as friction (§7) rather than asserted either way.

### 3.2 Live-smoke-testable (needs a real tool call or a cheap real agent turn, but not a full debate)

- **The filename-pattern write-block.** Verified via `ideas/backlog.md` as "CONFIRMED as a hard,
  report.md-specific tool error (forensics: `is_error: True`, 'Subagents should return findings
  as text, not write report files')" and independently re-triggered on `red/findings.md` in run
  2.[^WriteBlock] Grepping the repo's own hooks (`plugins/prosthetic-conscience/hooks/`) for any
  matching guardrail turns up nothing[^HookGrep] — **this is confirmed as a platform/environment
  restriction, not a hook this repo owns**, which matters for classification: it cannot be
  fixed by changing this repo's code logic (not caller-side plumbing, so not simulator-testable
  the way R1-HARNESS-1 is), but it also does not require a full agent debate to observe — a
  single Write-tool call to a `blue/report.md`-shaped path, with no LLM reasoning attached, either
  triggers the block or does not. Cheap, live, tool-level — the clean example of the missing
  middle tier between "pure logic" and "full production run."
- **ENAMETOOLONG on large Bash heredocs (Windows).** A real OS command-length limit, triggered
  once already (red-merge-r1, run 2) writing a ~236-line file in one heredoc spawn, forcing a
  6-call chunked workaround.[^Enametoolong] Testable cheaply and live (a single large heredoc
  write, no model reasoning needed) but not reproducible in a pure Node simulator unless the
  simulator also fakes OS-level command-length enforcement, which is not worth building — accept
  as live-smoke-testable, not unit-testable.
- **`/research --smoke` full-pipeline check** (already proposed in the backlog: 1 lane, 1 round,
  1 citation pass, cheap model, trivial topic).[^Backlog] This is the correct level for
  verifying that a real red-PASS gate and a real blue-synthesis actually fire correctly
  end-to-end — genuinely needs live model calls, genuinely doesn't need full-scale citation
  counts or multiple lanes to prove the wiring works.

### 3.3 Only-observable-in-production (real network state, real quota limits, real vendor content — no cheap substitute)

- **Leaf-node citation drift**: live sources changing after the fact (mem0's pivot to
  single-pass ADD-only, GitHub issue status flips, star-count changes) is inherently a
  point-in-time fact about the external world; the fix is procedural (record access-date deltas
  explicitly, as `run2-friction.md`'s red-merge-r1 already recommends[^DriftFriction]), not a
  code defect to eliminate.
- **PDF/table-extraction fidelity** and **primary-security-advisory access** — tool/environment
  capability gaps, not code defects (ranked in §4).
- **The quota-wall *trigger condition*** that produces a null `agent()` return is itself only
  observable in a real, expensive, at-scale run — but note the important asymmetry: the
  *trigger* is production-only while the *code's response to it* (the null-guard) is
  zero-token-unit-testable (§3.1). Classifying "the null-return defect" as a single bucket item
  is imprecise; the trimodal split should be applied per-defect-half, not per-defect-name, when
  a defect has both a rare live trigger and a testable code response.
- **Auto Dream / server-side-flag feature availability** (from the *subject* run's own findings,
  not FEOV's own code, but structurally the same class of "unobservable until a vendor flips a
  flag" defect) — noted here only as a cross-run pattern-match, not re-litigated.

---

## 4. Concrete simulator design (answers the report's explicit ask)

**Location:** `plugins/frank-exchange-of-views/skills/research-protocol/scripts/workflow.sim.test.js`
(pending the backlog's pending rename of `workflow.js` → `debate.js`[^Rename] — name the test
file to match whichever lands first; do not let the rename block the simulator or vice versa).

**No new dependency.** The repo currently has zero `package.json` and zero JS test
infrastructure anywhere[^NoPackageJson] — Node's built-in `node:test` + `node:assert` runs via
`node --test` with no install step, which is the lowest-complexity option and matches the
existing Go-tools precedent of plain-standard-library `_test.go` files with no external test
framework.[^GoTests]

**Mechanism.** `workflow.js`'s body references `args`, `agent`, `parallel`, `phase`, and `log` as
ambient bindings supplied by the Workflow tool at call time, not as imports — so the simulator
must construct the same execution context, not `require()` the file directly. Wrap the script
source in a function constructor (or `vm.Script`) that receives:
- `args`: the canned scenario input (`{ topic, runDir, lanes, maxRounds }` or a deliberately
  wrong-shaped value for the stringified/unbound-args cases).
- `agent(prompt, opts)`: a stub that (a) records every `(prompt, opts)` call for assertion — this
  is how the simulator catches "undefined" leaking into a prompt string without needing a real
  model call — and (b) returns the next canned envelope from a per-scenario queue keyed by
  `opts.label`.
- `parallel(thunks)`: `Promise.all(thunks.map(t => t()))` — two lines, faithful to the real
  harness's concurrency contract without needing real concurrency.
- `phase`, `log`: no-ops that push to an inspectable transcript array (needed for the §3.1
  "0-round message" assertion).

**Founding regression suite — beyond stringified-args and null-agent-returns (already named in
the backlog), add:**

1. Clean PASS in round 1 (no revision loop exercised at all).
2. Multi-round FAIL → revision → PASS (verify `blueEnv`/`redEnv` state threading between rounds).
3. Contested docket → judge → `deadlock: false` → loop continues.
4. Contested docket → judge → `deadlock: true` → loop halts with `deadlocked` flag set.
5. Safety ceiling (`maxRounds` hit while still `FAIL`) → `exhausted` flag set, verdict message
   distinguishes this from a judged deadlock.
6. Stringified `args` (JSON string instead of object) → assert no prompt contains `"undefined"`.
7. Unbound/missing `args` keys (object present, `topic`/`runDir` absent) → same assertion,
   distinct code path from #6.
8. `agent()` returns `null` (quota-wall simulation) at each of the loop's four call sites
   (red-lens, red-merge, judge, blue-respond) — four separate cases, not one, since a null return
   at each site crashes on a different subsequent dereference.
9. **Empty-candidates cascade**: lens-pass stubs return `null`/blocked envelopes → red-merge is
   still dispatched against an empty `red/candidates/` — assert the workflow either detects and
   short-circuits, or (if it currently doesn't) this test documents the gap as a known-failing
   regression until fixed (§3.1).
10. **Gap-id rollover**: a gap id present in round 1's `redEnv.gaps`, absent in round 2, present
    again in round 3 with the same id — assert whether round 3 correctly classifies it
    `contested` (recurrence) rather than `new` (currently does not; document as known-failing
    until `prevGapIds` is widened to full adjudicated history, not just the immediately prior
    round).
11. `--lanes 0` and `--lanes 1` — assert no crash, and separately assert (or document as a gap)
    whether a floor guard belongs in the caller (§2 recommends assigning method-diversity per
    lane over raising the floor, but the code should not crash on the edge either way).
12. `--maxRounds 0` — assert the emitted `log()` line distinguishes "never ran" from "ran and
    failed at round 0" (§3.1).
13. Citation-pass-count boundaries: `claim_count` of `0`, missing (default-20 path), `40`, `41`,
    and `500+` (verify the cap at 4 holds).
14. Malformed-but-non-null envelope (e.g., `redEnv` present but missing `gaps`) — document current
    behavior (likely an uncaught throw) as a known gap pending clarification of the Workflow
    tool's own schema-enforcement guarantee (§3.1, flagged as friction in §7).

This is 14 named cases against the backlog's existing 7 (clean PASS, FAIL→revision, contested
docket, deadlock, ceiling, stringified args, unbound args) plus null-returns and the already-fixed
`redEnv` null crash — roughly doubling the founding suite, all identifiable directly from the
run-1/run-2 defect corpus without inventing hypothetical failure modes.

---

## 5. H4 breadth — pre-run-4 changes, graded likelihood × impact × complexity

| # | Change | Likelihood (of the failure mode recurring) | Impact | Complexity | Grade |
|---|---|---|---|---|---|
| 1 | **Claim manifest, cheap half only** (tag single-lane-sourced claims at synthesis; full claim→citation→confidence ledger deferred) — fixes H2 (§1.4) | High (every multi-lane run) | High (restores minority-report signal to red's grading) | Low (set-difference over section content at a synthesis step that already exists) | **Build now** |
| 2 | **Per-lane assigned method/lens**, not just per-lane assigned starting hypothesis (§2) — e.g. one lane always runs a local-repo critical-stance pass | Medium (diversity gain is real but plateaus per the literature[^DiminishingReturns]) | Medium-High (produced 2 of the run's highest-value minority findings this run alone) | Low (one added sentence per lane-dispatch prompt) | **Build now** |
| 3 | **Workflow simulator** (§4) — founding suite of 14+ cases | High (every one of these has already occurred at least once) | High (would have caught the run-1 args bug before 253k tokens were spent[^Backlog]) | Low-Medium (no new dependency; ~1-2 days to wrap the script and write the 14 cases) | **Build now** |
| 4 | **Pre-create declared blackboard artifacts** (`blue/report.md`, `red/findings.md`, etc.) so subagents only append/edit, sidestepping the filename write-block (§3.2) — red's own recommendation | High (recurred on 2 distinct filenames in run 2 alone) | Medium (currently worked around via Bash heredoc, so not blocking, but adds token overhead every occurrence) | Low (the workflow script already creates the run directory in step 1 of `commands/research.md`; extend it to touch the known filenames) | **Build now** |
| 5 | **Rename blackboard artifacts away from the trigger pattern** (`blue/report.md` → e.g. `blue/corpus.md`) as backlog's alternate fix for the same write-block | Same as #4 | Same as #4, but touches every template/prompt/reference to the filename across the plugin | Medium (higher blast radius than #4 — every doc, prompt, and cross-reference to `report.md`/`findings.md` needs updating) | **Prefer #4; risk-accept #5 as unnecessary** — pre-creating the file is strictly cheaper than a plugin-wide rename for the identical fix outcome (subagents append instead of create); recommend closing this backlog item as superseded by #4 rather than doing both |
| 6 | **PDF table/full-text extraction** — ranked #1 by independently-reported friction volume (§6) | High (recurred every round of run 2, both roles) | High (kept 3+ figures at "unable-to-corroborate" across 4 rounds — R1-19, R1-28, R3-14/15) | **Lower than the backlog assumed**: off-the-shelf MCP servers already do this (`arxiv-latex-mcp` fetches LaTeX source for exact figures; `pdf-reader-mcp` extracts tables with cell data, bounding boxes, and confidence scores) — no bespoke `sc-pdf-extract` Go tool needed, contrary to the backlog's own tentative candidate.[^PdfMcp] | **Build now, via adoption not construction** — re-grade from the backlog's implicit "Medium/High complexity, unscoped" to **Low complexity** once scoped as "wire an existing MCP server," which changes the priority ordering materially (see §6) |
| 7 | **Primary Anthropic security-advisory access** | Medium (recurred rounds 2-3, not round 4 — see §6) | High when it fires (was load-bearing for the R2-2 double-bind keystone) | Unknown/possibly not fixable from this side (may require an authenticated or first-party channel this plugin cannot grant itself) | **Risk-accept for now, revisit if it becomes load-bearing again** — round 4's friction record shows the team successfully engineered around it (unconditional de-authorized channel, §13.7(4) in the subject run) rather than needing the access; the workaround already generalizes, so building bespoke access infrastructure for a gap the team already designed around fails blue's own complexity-cost test (§7 makes the case explicitly) |
| 8 | **Access-date-delta recording** for citations (protocol-level, not code) — mitigates live-source drift (§3.3) | High (already observed 4+ times: mem0 pivot, issue-status flips, star counts, arXiv figure corrections) | Medium (drift is usually caught by re-verification, not silently missed — the cost is re-work, not undetected error) | Low (a footnote-template field addition) | **Build now** |
| 9 | Formalize trajectory capture (`journal.jsonl` → `<run>/trajectories/`, gzip transcripts) — already in both backlog copies | High (currently evaporates every run) | Medium (this retrospective itself needed it and it was present only because of a manual capture step) | Low (a copy-and-gzip step after the workflow returns) | **Build now** — this retrospective is itself the demonstration of its value: §1, §3 above could not have been written without `journal.jsonl` |

**Explicitly risk-accepted, not built (pragmatist duty against scope creep):**

- **Full claim manifest** (citation + self-graded confidence + lane provenance for *every*
  claim, not just single-lane ones) — deferred per §1.4's own complexity argument until the
  cheap half proves insufficient in practice.
- **Raising the default lane count** (2→3+) as a blanket policy — the diminishing-returns
  literature[^DiminishingReturns] and §2's method-vs-headcount analysis argue against this as the
  first lever; #2 above is the cheaper, better-targeted fix for the same symptom.
- **Round-scoped audit narrowing** (backlog item: later rounds re-audit only changed sections +
  contested gaps + spot checks, trading against the "always re-read the full report" principle)
  — flagged by the backlog as needing "the human's call" precisely because it trades against a
  named design principle; not this retrospective's call to make unilaterally, and multi-agent
  literature on diminishing returns from added inspection rounds supports at least considering it,
  but changing a stated protocol invariant is a bigger decision than the other items here and
  should go through its own review, not ride in as a side-effect of this retrospective.

---

## 6. H5 breadth — ranking friction by distinct-agent-role count (confirms, refines the frontier's prediction)

Counting distinct-role attributions directly in `run2-friction.md` (35 entries across 4
rounds):[^FrictionCount]

| Gap | Roles reporting it, by round | Distinct roles | Rounds present |
|---|---|---|---|
| **PDF full-text / table-extraction** | red (r1,r2,r3,r4), blue (r1,r2,r3,r4), judge (r2) | **3 (red, blue, judge)** | **all 4 rounds** |
| **Primary Anthropic security-advisory access (CVE-2026-21852)** | red (r2,r3), judge (r2), blue (r2,r3) | **3 (red, judge, blue)** | rounds 2-3 only, **not round 1 or round 4** |
| Filename write-block heuristic | red (r1) | 1 | round 1 only |
| ENAMETOOLONG heredoc chunking | red (r1) | 1 | round 1 only |
| Live-source drift (access-date deltas) | red (r1) | 1 | round 1 only (but generalizes — see §3.3) |
| Trajectory-extractor implementation opacity | red (r3) | 1 | round 3 only |
| No sandbox to test Auto Dream's runtime behavior | red (r4) | 1 | round 4 only |
| Springer/auth-wall unfollowable primary source | red (r4) | 1 | round 4 only |

This **confirms** the frontier's H5 core claim: two gaps dominate by both role-count (3 distinct
roles each) and persistence (present across most/all rounds), while everything else is a
round-1-or-single-round one-off. It **refines** the frontier's prediction on one point: H5
guessed the write-block/ENAMETOOLONG/preflight-guard complaints "cluster in round 1 only (already
fixed)" — true for run 2's ENAMETOOLONG and write-block, but the run-1 harness-defect record shows
the *preflight-guard* complaint (uninitialized run-directory) was reported by nearly every one of
16 agents in that separate, earlier run before being fixed[^Journal] — meaning the *stable,
currently-unresolved* list is exactly the two H5 predicted (PDF-extraction, primary-advisory
access), but the *historically loudest* single complaint (by raw agent count, now resolved) was
the args-binding bug, which is worth keeping in the record precisely because it demonstrates the
fix-and-it-stops pattern the other two gaps have not (yet) followed.

**The round-scoped difference between the two dominant gaps is itself informative and was not in
the frontier's prediction**: PDF-extraction friction is reported in *every* round including round
4, while the CVE-advisory friction *stops* appearing in round 4 even though the underlying
capability gap was never closed. Cross-referencing `blue/CHANGELOG.md`'s Round 2 entry — "§13.7
R1-8/R2-2 lead docket CLOSED ... double-bind resolved by UNCONDITIONAL de-authorized channel"[^ChangelogR2]
— shows why: the team engineered around the CVE-advisory gap (making the recommendation robust to
either factual branch) rather than needing the access, so agents stopped hitting it as a live
blocker. PDF-extraction has no equivalent workaround — every hedge ("approximate," "attributed,"
"unable-to-corroborate-at-leaf-node") is a standing cost paid every round, which is exactly why
it, and not the CVE gap, should be item 6 in §5's build list rather than item 7.

---

## Footnotes

[^ChangelogR0]: `research/2026-07-12_memory-architecture/blue/CHANGELOG.md`, Round 0 entry. Local file, accessed 2026-07-13.
[^BlueReportGrep]: Grep `-i "lane"` against `research/2026-07-12_memory-architecture/blue/report.md` (2972 lines) — 7 total matches (5 shown after excluding the §10/§8 header duplicates), all method-level or section-header, none per-claim. Local verification, accessed 2026-07-13.
[^BlueReportUnverified]: `research/2026-07-12_memory-architecture/blue/report.md` §10 (Unverified items): "cited by the proposal without independent corroboration in either lane." Local file, accessed 2026-07-13.
[^RedFindingsGrep]: Grep `-i "lane"` against `research/2026-07-12_memory-architecture/red/findings.md` (695 lines) — 1 match, line 156: "Disconfirming budget met in both blue lanes. Not a gap." Local verification, accessed 2026-07-13.
[^RedAuditorSpec]: `plugins/frank-exchange-of-views/agents/red-auditor.md`. Local file, accessed 2026-07-13.
[^R2-10]: `research/2026-07-12_memory-architecture/red/findings.md`, lines ~424-451 (`[^SingleUserLowRisk]` footnote-provenance finding, R2/R3 status). Local file, accessed 2026-07-13.
[^Lane1]: `research/2026-07-12_memory-architecture/blue/candidates/lane-1.md` (355 lines, H1-deep). Local file, accessed 2026-07-13.
[^Lane2]: `research/2026-07-12_memory-architecture/blue/candidates/lane-2.md` (321 lines, H2-deep). Local file, accessed 2026-07-13.
[^ClaimManifest]: `ideas/backlog.md`, item "(5) CLAIM MANIFEST." Local file, this repo's live (non-frozen) backlog, accessed 2026-07-13.
[^ResearchCmd]: `plugins/frank-exchange-of-views/commands/research.md`, `--lanes` flag documentation ("default 3"). Local file, accessed 2026-07-13.
[^DiminishingReturns]: Multi-agent debate diversity/scaling literature survey (search results synthesizing arXiv 2603.20640 "Efficient Multi-Agent Debate via Diversity-Aware..."; arXiv 2601.19921 "Demystifying Multi-Agent Debate: The Role of Confidence and Diversity"; VentureBeat "'More agents' isn't a reliable path to better enterprise AI systems"; arXiv 2605.00914 "The Cost of Consensus"), accessed 2026-07-13.
[^Workflow]: `plugins/frank-exchange-of-views/skills/research-protocol/scripts/workflow.js`, line 16 (`const { topic, runDir, lanes = 3, maxRounds = 12 } = args`). Local file, accessed 2026-07-13.
[^Journal]: `research/2026-07-12_feov-retrospective/inputs/run1-defect-record/trajectories/journal.jsonl` (32 lines, 16 agent dispatches, all independently detecting the uninitialized run-directory/topic defect). Local file, accessed 2026-07-13.
[^Backlog]: `ideas/backlog.md`, items on the workflow simulator, the `redEnv` null-crash founding case, and the `/research --smoke` proposal. Local file, accessed 2026-07-13.
[^RedMergeR2]: `research/2026-07-12_feov-retrospective/inputs/run1-friction.md`, line 7 ("No round-2 red candidate lens passes were ever produced, so the merge had no inputs by construction"). Local file, accessed 2026-07-13.
[^WriteBlock]: `ideas/backlog.md`, item "the workflow-subagent write-block is CONFIRMED as a hard, report.md-specific tool error." Local file, accessed 2026-07-13.
[^HookGrep]: Grep `-i "report|findings"` against `plugins/prosthetic-conscience/hooks/` — no matches; the write-block is not implemented by this repo's own hooks. Local verification, accessed 2026-07-13.
[^Enametoolong]: `research/2026-07-12_feov-retrospective/inputs/run2-friction.md`, red-merge-r1 entry. Local file, accessed 2026-07-13.
[^DriftFriction]: `research/2026-07-12_feov-retrospective/inputs/run2-friction.md`, red-merge-r1 entry on live-source drift and access-date deltas. Local file, accessed 2026-07-13.
[^Rename]: `ideas/backlog.md`, item "rename `skills/research-protocol/scripts/workflow.js` → `debate.js`." Local file, accessed 2026-07-13.
[^NoPackageJson]: Glob `**/package.json` across the repo — no results. Local verification, accessed 2026-07-13.
[^GoTests]: `plugins/prosthetic-conscience/tools/cmd/*/main_test.go`, `tools/internal/*/*_test.go` — standard-library `testing` package, no external framework. Local verification, accessed 2026-07-13.
[^PdfMcp]: "GitHub - takashiishida/arxiv-latex-mcp" and "GitHub - SylphxAI/pdf-reader-mcp" (table extraction with cell data, bounding boxes, confidence scores), accessed 2026-07-13.
[^FrictionCount]: `research/2026-07-12_feov-retrospective/inputs/run2-friction.md`, full 35-entry count by role and round. Local file, accessed 2026-07-13.
[^ChangelogR2]: `research/2026-07-12_memory-architecture/blue/CHANGELOG.md`, Round 2 entry, §13.7 closure. Local file, accessed 2026-07-13.
