# How should sleeper-service, the autonomous learning loop (Phase 4), be designed? — research report

**Verdict:** UNVERIFIED (after 5 debate rounds — terminated at the round safety ceiling with 10 gaps carried; red's standing verdict is FAIL, the lead's deadlock check was FALSE) · **Run:** `research/2026-07-17_sleeper-service-design/`

**TL;DR:** The design is implementable and fully specified: a wrapper-driven nightly loop that mines the suite's durable artifacts (friction, cost, board telemetry, run records, backlog/doubts, red's gap-pattern memory), ranks deterministically at zero tokens, researches ONE pick at smoke scale, and writes only `research/` + `ideas/` — with promotion into rules/skills human-gated behind seven structural layers, at least one always outside the model's write reach. Five red rounds converged monotonically (gap mass 118.75 → 81.5 → 55.0 → 46.0 → 30.0; open count 30 → 23 → 17 → 16 → 10) and every external evidence base held at the leaf; the failures were composition, and each round's board forced a structural invariant now in the design (checked-record gates, all-paths guard totality, bare-`Bash` tool-boundary close, positively-defined corroboration pool). The verdict is UNVERIFIED because the round ceiling fired before red could audit blue's round-5 revision: the 10 carried gaps' repairs are recorded and claimed, not red-verified. Sharpest caveat: the two round-5 mediums (seat-agent Bash boundary R5-1; corroboration-pool poisoning R5-3) were closed by stated invariants in that unaudited revision, and two named build-time probes — OQ23(e) (headless seat `--settings` inheritance) and OQ18(c)(3) (hook allowlist attached-form flags) — plus the standing OQ1/OQ2 headless probes are the Phase-4 acceptance legs the build PR must run.

## Verdict detail: outstanding gaps at close, dispositions, and the compromise rationale

**Compromise rationale.** The gate never soft-passes, and it did not here: red's last issued verdict is FAIL (round 5: 10 open residues, mass 30.0, max severity medium). The round-5 LEAD adjudication carried all 10 gaps with owed direction and ruled deadlock FALSE ("the debate continues into round 6"); the run's 5-round safety ceiling then terminated the debate before any round-6 red pass could audit blue's round-5 revision. Blue executed every owed direction (per-gap edits and propagation greps in `blue/CHANGELOG.md` Round 5) — but those repairs are blue-claimed, not red-verified. The honest stamp is therefore UNVERIFIED: the full five-round-hardened design is delivered, with the round-5 revision explicitly labeled unaudited. The next build PR inherits (a) a re-audit of the CHANGELOG Round 5 edit set and (b) the named build-time probes below as acceptance tests.

**Outstanding gaps (all 10 ruled `carried` by the round-5 LEAD; blue's round-5 repair applied; UNAUDITED at close):**

| Gap | Severity | Finding (short) | Round-5 repair of record (unaudited) |
|---|---|---|---|
| R5-1 | medium | §4.3 layer 4 (iii) vs §4.2 contradict on whether nightly FEOV seat agents sit inside the bare-`Bash`-deny boundary; both horns defective; layer-1 `--settings` inheritance for headless-spawned seats never probed | Horn picked: seats BOUND by design requirement; sleeper-guard hook generalized to deny the whole Bash channel for sleeper-marked runs (inheritance-independent close); seat-Bash capability cost owned; dead `Bash(git …)` allow rules re-labeled; refuted "auto-approves read-only git regardless" comment corrected; OQ23(e) named probe |
| R5-2 | low-medium | §3.4 rung-0 cell + §3.3 Phase-4 acceptance test still specify the pre-trampoline payload shape; test unsatisfiable as written, cheapest pass undoes R4-1 | Both sites rewritten to the trampoline shape; acceptance test restated two-legged: wrapper `--manual` produces the run dir; `claude -p "/self-improve"` produces NONE (trampoline inertness verifiable) |
| R5-3 | medium | Corroboration pool defined NEGATIVELY; three rounds of per-surface patches; rung-2 Desktop tasks poison the shared corpus AUTOMATICALLY behind an instructional-only gate | Root invariant stated once (the R4-9 move): pool POSITIVELY defined — a dir corroborates only with affirmative non-sleeper provenance (pre-epoch date or interactive-origin marker stamped at creation); anything unattributable QUARANTINED counts-for-nothing, fail-closed; OQ24 promoted from deferred to built |
| R5-4 | low | R4-6 confinement re-introduces dir-NAME keying against §1.5's own doctrine; the sub-run slug is unknowable after a hard-kill | Doctrine qualified (name-keying confines retroactive-uncertain sweeps, never assigns origin); wrapper logs every sub-run dir PATH at creation so confinement matches recorded paths |
| R5-5 | low | ATTACHED `-o<value>` form escapes the `Bash(* -o *)` belt — the enumerate-and-extend regress recurring inside the R4-2 repair itself | Belt declared KNOWN-incomplete on attached short forms, not extended again; hook read-allowlist named the close of record; `git format-patch -o<attached>` added to OQ18(c)'s probe matrix |
| R5-6 | low | `expected-memory-deny` keyed on target path, not actor/shape — memory-poisoning ATTEMPTS file into the ignored bucket at any volume | One predicate: the bucket admits only the nightly red-merge seat's own bounded write shape; anything else surfaces as `unexpected-memory-write-attempt`, persisting like TAMPER |
| R5-7 | low | Sleeper's own hooks.json fires in every interactive session with bootstrap-guard coverage unstated; flag-ABSENT anomaly nags never-configured installs | hooks.json wraps every command in the bootstrap guard; §6 row 9 re-pointed; empty-bin acceptance check (OQ10) extended; flag-missing anomaly gated on a prior successful-run record |
| R5-8 | low | §0 tree ("the implementable shape") omits the fourth code artifact and ALL hook registration — a builder ships an unregistered enforcement hook | Tree redrawn: `hooks/hooks.json` (registering the sleeper-guard PreToolUse fence AND the SessionStart staleness hook) + the SessionStart executable |
| R5-9 | low | `rejected` re-surface arms key on a rejection DATE no artifact records; `regression` token domain unpinned | Dated status field (`status: rejected 2026-07-17`) both arms key on; `regression` pinned docket-only; `rejected-recurring` setter stated |
| R5-10 | trivial | est_complexity "NAMED source" unparseable and empty against the actual pinned corpus — divisor universally inert while text implies it activates | Literal `cx:<1-5>` token convention stated; field noted currently unpopulated at the pin (forward-looking curation surface) |

**Adjudicated earlier (out of red's verdict pool, recorded here per the never-dropped rule):**

- **R1-7 — risk_accepted (LEAD, round 2; scope: this run only).** `inputs/PINNED.md` asserts `plans/claude-port-plan.md` at pin `7bc501e`, where the path does not exist — a run-infrastructure defect, not blue's. Consequence: every port-plan citation in this report is **snapshot-grade** (verified against working tree `6df52af`), not pin-grade. The debt is the lead's: setup tooling must validate every pinned path (`git cat-file -e <pin>:<path>`) at run creation and fail loudly, or stage cross-corpus artifacts into `inputs/` at pre-create. The defect recurs in every run pinning a cross-repo path until fixed.

## The Catechism

The agreed post-debate positions. (Assembly note, disclosed: blue's report template carries no Catechism section, so these seven answers are assembled by the judge from the debated text — the answers cite the sections where the debate DID audit the underlying claims, but the catechism *as a unit* was never itself a red target.)

1. **What are we trying to do?** Build sleeper-service: a nightly, unattended loop that reads the suite's own accumulated exhaust — friction records from every run, the cost ledger, board telemetry, run-report open questions, `ideas/backlog.md` + `doubts.md`, red's gap-pattern memory — picks exactly ONE improvement candidate by deterministic ranking, researches it at bounded smoke scale, and leaves a human-reviewable idea stub with provenance and 3–5 alternatives. It writes only `research/` and `ideas/`. Promotion into rules/skills happens only through `/graduate`: interactive, full-strength FEOV debate, and a human-merged PR — never scheduled, never autonomous.

2. **How is it handled today, and what does that cost us?** Manually: insights survive only if the operator notices and files them. The pinned friction harvests show the same complaints recurring across runs unactioned; the backlog (25 items at the pin) accretes without curation or staleness pressure. The cost is silent signal rot — captured-but-never-triaged friction is the dominant waste class in the suite's own telemetry (§1.1–§1.3 of the blue report).

3. **What is new here, and why do we believe it works?** Artifact-mining, never introspection — the external literature (Reflexion, Voyager, DGM, SICA, STOP, and the negative result on intrinsic self-correction) plus both pinned harvests support consuming durable, pre-curated artifacts over asking a model to recall its failures (§1). Mechanics are cheapened to zero tokens (a deterministic Node harvester); judgment is never cheapened (one pick, full-strength model, smoke-scale FEOV). Headless viability is proven by live probe on this box, not inferred: a plugin command executed under `claude -p` on CLI 2.1.212, write denied and honestly reported (§3.1).

4. **The case against — at full strength.** (a) Autonomous loops editing their own constraints is a *realized* failure class, not a hypothetical: AI Scientist modified its own execution script; DGM agents removed their guardrails; STOP measured 0.42% unsafe self-modification even with warnings — and permission rules alone are empirically leaky (#22055 closed NOT PLANNED). The naive version of this tool is unsafe by the report's own evidence, which is why seven gate layers exist. (b) The loop authors its own future harvest inputs: self-poisoning of the ranking corpus took THREE rounds of red pressure (R2-5/R2-6 → R3-3/R3-5 → R4-1/R4-6/R4-13 → R5-3) to reach a root fix, and that fix is unaudited. (c) Proposal fatigue at the human gate is the documented Dependabot failure mode; a nightly stub generator is exactly that shape. (d) Daily smoke-scale stubs may be too shallow to ever graduate (H2's named falsifier). (e) Unattended spend on a harness with NO programmatic quota pre-check (confirmed absent at this auth tier) — guards are honest but static. (f) The headless surface is fragile: the open MCP-headless bug trio, the `--bare` default-flip risk, plugin-cache empty-bin windows, workspace-trust preconditions. (g) Maintenance: four code artifacts, two cross-plugin edits, two operator-owned configs — every one a surface that can silently drift, and this debate's five rounds of composition defects are the measure of how easily specified guards under-reach. (h) The run's own meta-lesson: believing an unaudited guard is the exact error the verdict stamp refuses.

5. **Of interest, or merely interesting?** Of interest — conditionally. It displaces almost nothing: unattended nightly runs at ~$2–5 under a ~$50 monthly ledger, zero operator attention except the morning stub review, and the null alternative (keep the loop manual) was priced against the suite's own evidence that manual triage demonstrably loses signal (R1-14, answered in §1.4). The condition: the gates must hold as specified. If the build-time probes fail (OQ1/OQ2/OQ23(e)/OQ18(c)(3)), the loop stays at rung 0 (manual wrapper invocation) — still useful, no longer autonomous — rather than running unattended on eroded gates.

6. **What changes if it works — and if we simply don't do it?** Works: recurring friction gets triaged nightly into auditable, provenance-carrying stubs; the backlog gains staleness pressure and re-surface timers; graduation feeds FEOV with pre-researched candidates; the suite's self-observation becomes cumulative instead of episodic. Don't: nothing breaks — the suite continues exactly as today, and signal continues rotting silently. That asymmetry (loop expendable, gates not) is priced INTO the design: every safety tradeoff resolves conservative because a skipped night costs nothing (risk row 6, risk-accepted by design).

7. **What does it cost, and where would we stop?** Build: four small code artifacts (harvest.mjs, sleeper-wrapper.mjs, sleeper-guard, SessionStart staleness hook + hooks.json), two command prompts, one skill file, one scheduling doc, two operator-owned configs, two one-line cross-plugin edits. Run: per-run `--max-budget-usd` + `--max-turns` + operator-owned month-to-date ledger preflight, fail-closed. Stop conditions are named and mechanical: resume cap k=3 → dead-mark; per-cause HALT after M=3 same-signature deaths; dead-man surfacing via `/doctor` + SessionStart warning at N=7 days. Abandon checkpoints: stub→graduation survival rate too low → switch to weekly-full-strength or retire the loop (risk row 2); any TAMPER flag or hook-log guardrail-targeting attempt → halt and human review; failed acceptance probes → the rung stays manual. The Phase-4 acceptance test is two-legged (wrapper `--manual` produces a run dir touching only `research/`+`ideas/`; `claude -p "/self-improve"` produces none).

## Technical foundations

Established at the leaf (full citations in the embedded blue report's merged footnote namespace; ~100+ statement↔reference pairs independently leaf-verified per round in `red/citation-ledger.md`; ten living-source claim sets re-fetched zero-drift on 2026-07-17 in round 5):

- **Harness, verified live on this box (CLI 2.1.212):** plugin commands execute under `claude -p`; a denied write is honestly reported (blue §3.1, lane-3 probes). `--input-format stream-json` is present at the pinned CLI (`claude --help`, round 2), grounding the phase-driven wrapper drive. The dontAsk read-only Bash carve-out and its deny-supremacy semantics are doc-verified live, including the bare-tool-name deny removing the tool entirely (blue §3.2/§4.2; permissions doc quoted verbatim round 3).
- **Harness limits, doc- and issue-grounded:** no programmatic spend-limit API; Rate Limits API unreachable at this auth tier (§5.1). MCP-headless flake is an open bug class (#76239, #68375 — OPEN re-confirmed 2026-07-17; #68375 now carries a `stale` label). Headless plugin expansion carries the `--bare` flip risk (#837 class, CLOSED COMPLETED). The 10-minute background-wait ceiling and workspace-trust preconditions are named traps (§3.3).
- **The write gadget record, reproduced not assumed:** `git log --output` (R3-15), `git format-patch -o <path>` spaced form (R4-2), and the ATTACHED `-o<path>` form (R5-5) each leaf-reproduced on this box at lens, merge, and judge seats — exit 0, arbitrary out-of-repo writes, no prompt. This record is why the close of record is the bare `Bash` deny + hook read-allowlist, not pattern enumeration.
- **External literature (primary sources):** Reflexion/Voyager/DGM/SICA/STOP corpus + the negative result on intrinsic self-correction ground artifact-mining over introspection (§1); AI Scientist's sandbox escape, DGM guardrail removal, and STOP's measured 0.42% unsafe self-edit rate ground the treat-per-attempt-probability-as-nonzero-forever posture (§4); Dependabot base rates ground the proposal-fatigue mitigation (§1.4/§2.3).
- **Pinned internal artifacts (evidence base at `7bc501e`):** ideas/backlog.md (25 items), doubts.md, both friction harvests, the efficiency-investigation run record, red's gap-pattern memory (1,557-line mirror, byte-verified). **Exception:** `plans/claude-port-plan.md` does not exist at the pin (R1-7) — port-plan quotes are snapshot-grade at working tree `6df52af`.
- **Pricing and cost figures:** the pricing grid leaf-verified HIGH (round 1) and re-fetched zero-drift twice (rounds 2, 5); the Batch ≤24h expiry sub-claim resolved at the batch-processing page; cap/HALT arithmetic survives three independent recomputes (round 5).

## Analysis

The argument runs: all five frontier hypotheses came back SUPPORTED, three with conditions the design now carries explicitly (blue Round 0). What the debate then did — and this is the run's finding as much as the design is — was expose, five rounds running, that **verified pieces do not compose for free**:

- **Round 1** (30 gaps): the evidence base was "unusually sound"; the failures were composition — a write-capable subprocess channel contradicting the read-only claim (R1-16), the loop authoring its own future harvest inputs (R1-25), a gate stack that silently didn't exist at rungs 2–4 (R1-27), a fence failing OPEN on marker loss (R1-28).
- **Round 2** (23): the round-1 *repairs* landed but the repair MACHINERY didn't verify itself — a canary certifying the wrong thing (R2-1/R2-2), same-round fixes that don't compose (R2-3), evidence with no watchman (R2-7/R2-8/R2-9). Answer: **invariant 7** — every wrapper gate emits a checked, human-surfaced liveness/outcome record.
- **Round 3** (17): repairs held for the scheduled/completed happy path but not the default mode, abort path, or sibling surface (R3-2 the archetype). Answer: **invariant 8** — every guard specified over ALL exit paths, authorship surfaces, and execution modes.
- **Round 4** (16): round-3 repairs closed the named path but not its sibling. Answer: **structural closes** — the bare `Bash` deny at the tool boundary, the git hook read-ALLOWLIST, the `/self-improve` trampoline, the R4-9 no-status-is-timer-free root invariant.
- **Round 5** (10): the structural closes were derived over the TOP-LEVEL session, not the surfaces where nightly work executes — the Workflow-spawned seat population (R5-1) and the negatively-defined corroboration pool (R5-3). Blue's (unaudited) answer applies the lead's remedy shape: state the invariant once over the TRUE surface — seats bound with an inheritance-independent hook close; the pool positively defined with fail-closed quarantine.

**Alternatives considered and beaten** (each argued in the embedded report, not merely dismissed): introspection-based consumption (refuted by the self-correction literature and both harvests, §1.1–§1.2); the permission-rules-only gate variant (REFUTED at the leaf — the design's own probe record shows rules alone leak, §4.1); an invocable payload-carrying `/self-improve` command (closed R4-1 — trampoline instead); a corroboration pool patched per-surface (closed R5-3 — root invariant instead); Console-scraping quota checks, cloud-VM-by-default, bespoke Windows sandboxing, a qmd daemon supervisor (rejected on likelihood×impact×complexity, recorded in §6's pragmatist scope defense).

**What is genuinely unresolved:** whether the round-5 revision holds under audit; whether debate.js-spawned seats inherit the headless `--settings` layer-1 profile (OQ23(e) — the design no longer depends on it, but the answer decides whether layer 1 also binds seats); the hook allowlist's handling of attached short-form flags (OQ18(c)(3)); the bare `--plugin-dir` headless recipe and headless hook-fence liveness (OQ1/OQ2). All are build-time acceptance probes, deliberately: they require BUILT artifacts a design-phase debate cannot execute.

## Risk matrix

Graded likelihood × impact × complexity-to-mitigate. Condensed from the debated §6 of the embedded blue report (full argued rows there — the authority for every disposition below), plus run-level rows the assembly owns. Risk-accepted items are recorded with rationale — elevated, never dropped.

| # | Risk | L | I | Cx | Disposition |
|---|---|---|---|---|---|
| 1 | Headless plugin-command expansion regresses / `--bare` becomes the `-p` default | Low–Med | High | Low | Mitigate: step-0 loaded-plugins assertion aborts loudly; two-legged Phase-4 acceptance test; version pin |
| 2 | Daily stubs too shallow to graduate (H2 falsifier) | Med | Med | Low | Mitigate: ledger stub→graduation survival metric; named revisit trigger (weekly-full-strength) |
| 3 | Proposal fatigue at the human gate | Med | Med | Low | Mitigate: 1-stub/run cap, dedupe, 30-day staleness sweep, R4-9 root invariant (no status timer-free) |
| 4 | Subprocess write outside boundary (no full Windows sandbox) | Low-Med | High | Med | Layered mitigation; residual **risk_accepted**, argued WITHOUT actor-benignity: bare `Bash` deny closes the session's subprocess surface structurally; surviving subprocess is pinned Workflow code under schema-bound args; every-exit snapshot compare detects in-repo tampering; full-sandbox complexity exceeds bounded L×I |
| 5 | No programmatic quota pre-check | Certain | Low | Low | **risk_accepted**: static ceilings suffice; a Console-scraping workaround is complexity strictly worse than the risk |
| 6 | Scheduler misses days (sleep/laptop) | High | Trivial | Low | **risk_accepted** by design: missed-run-tolerant trigger + idempotence; a skipped day costs nothing |
| 7 | Loop self-modifies its guardrails | Low per-attempt, compounding | Critical | Med | Mitigate: layers 1–7 (§4.3); STOP/DGM/AI-Scientist evidence says treat per-attempt probability as nonzero forever |
| 8 | MCP headless flake starves qmd recall (#76239/#68375) | Med | Low | Low | Mitigate: daemon `/health` preflight + `--strict-mcp-config` + Grep/Read degradation (proven workable) |
| 9 | Empty-bin window hook crash-storm — now BOTH hooks.json files (R5-7) | Med | Low | Low | Mitigate: bootstrap guard wraps both files' commands; OQ10 empty-bin acceptance check; residual noise accepted |
| 10 | Self-poisoning of the ranking input (edit + authorship paths) | Low-Med | Med | Low | Mitigate: positively-defined corroboration pool w/ fail-closed quarantine (R5-3, unaudited), wrapper-stamped origin markers, 1-per-class cap, severity-gated corroboration |
| 11 | Cloud Routine acts through operator identity, connectors default-included | Med | Med | Low | Mitigate (if that rung is adopted): trim connectors, keep `claude/`-branch restriction ON |
| 12 | Unattended overspend | Low | Med | Low | Mitigate: per-run `--max-budget-usd` + `--max-turns` + operator-owned ledger preflight, fail-closed |
| 13 | Read+egress exfiltration by an injected loop | Low-Med | Med-High | Low | Mitigate: Read/Grep/Glob repo-scoped, bare `Bash` deny, hook Bash-channel deny on the seat surface (R5-1, unaudited), WebSearch dropped; residual arXiv-WebFetch + local-qmd egress **risk_accepted**: single-operator box, repo largely public, channel narrow/enumerable/session-logged |
| 14 | Indirect prompt injection via retrieval surface (qmd corpus, arXiv uploads) | Med | Med | Low | Mitigate: untrusted-origin labels bar web-derived claims from ranking; origin cap bounds amplification; graduation re-derives every claim under adversarial review with the human present |
| 15 | Persistent loop death goes unnoticed | Med | Low-Med | Low | Mitigate: resume cap k=3, per-cause HALT M=3, doctor dead-man line + SessionStart staleness warning; detection-latency residual accepted (bounded by next interactive session; single-operator box) |
| 16 | Rung migration silently sheds gates (rungs 3–4) | Med | High | Low | Mitigate: per-rung gate-survival table; rung-3/4 adoption is a graduation-grade decision with its own stub + human gate, never a config toggle |
| 17 | **R1-7** — pin asserts nonexistent path; port-plan citations snapshot-grade only | Certain (defect verified present) | Low | Low | **risk_accepted** (LEAD, round 2; this run only): design conclusions do not rest on pin-grade port-plan text; fix owned by run infrastructure (pin validation / inputs-staging at run creation); recurs every cross-repo-pinning run until fixed |
| 18 | **Assembly row** — round-5 repairs unaudited (safety ceiling fired before a round-6 red pass) | Certain (structural fact of this run) | Med (two medium-severity closes among the unaudited set) | Low | Disposition: UNVERIFIED stamp; build PR must treat CHANGELOG Round 5 edits as claims requiring re-audit and run OQ1/OQ2/OQ23(e)/OQ18(c)(3) as acceptance probes |
| 19 | Pragmatist scope-defense rejections: Console-scraping quota checker; cloud-VM daily default; bespoke qmd supervisor; quota-introspection service; bespoke Windows sandboxing | — | — | High in every case | **risk_accepted** ×5: complexity exceeds likelihood×impact in each (argued under §6 of the embedded report) — recorded, not ignored |

## Blue team report (in full)

*`blue/report.md` embedded verbatim below — union, not summary. This is the living report as it stood at close: the round-0 three-lane synthesis plus five rounds of additive revision (change log in `blue/CHANGELOG.md`; the Round 5 edit set is unaudited per the verdict).*

---

# blue report — How should sleeper-service, the autonomous learning loop (Phase 4), be designed?

**Round 0 synthesis — union of three method-lens lane candidates:**
lane-1 (adversarial-disconfirming-first), lane-2 (primary-literature), lane-3
(local-repo critical-stance / live-probe). Merge is by inclusion: overlapping claims
deduplicated, no substantive content dropped. **Provenance convention:** claims appearing in
exactly ONE lane's draft carry a minority marker, e.g. `[minority: lane-3/local-probe]` —
these are minority reports red must weigh differently from convergent ones. Unmarked claims
appeared in two or more lanes. Footnote labels are merged where lanes cited the same source;
each merged footnote notes which lanes cited it. Evidence base pinned at `7bc501e`
(`inputs/PINNED.md`); external access dates 2026-07-17 throughout; living sources flagged.

**One cross-lane conflict is carried openly** (§2.2, step 6): lane-2's backlog-append step
vs lane-3's self-poisoning guard. Both positions are preserved with a proposed
reconciliation; this is a declared open design decision, not a silent resolution.

**Round 1 revision (2026-07-17):** all 30 red gaps addressed additively — per-gap edit log
and propagation greps in `blue/CHANGELOG.md` Round 1. Design deltas of record: harvest
moves wrapper-side, outside the model session (R1-16); the cost ledger moves to the
operator-owned dir (R1-19); the write-fence gains a step-0 denial canary (R1-28);
`--plugin-dir` is pinned to an operator-owned read-only plugin copy, never the working
tree (R1-15); the nightly profile drops WebSearch and scopes Read/Grep/Glob to the repo
(R1-17/R1-18); §1.5 covers authorship, not just edits (R1-25); wrapper start/end
snapshots restore durable tamper evidence (R1-26); §3.4 gains a per-rung gate-survival
table (R1-27) and a resume cap + dead-man surface (R1-29).

**Round 2 revision (2026-07-17):** all 22 round-2 gaps addressed (16 lead-carried
directions executed + 6 red new-mints) — design deltas of record: the wrapper becomes a
PHASE-DRIVEN stream-json session driver with a fired-record canary (R2-1/R2-2); the FEOV
execution locus is probe-determined MIXED and re-hosted (R2-3); origin tagging moves to
causal provenance markers (R2-5); corroboration is decided severity-gated (R2-6); the
snapshot delta gains its watchman (R2-7); the plugin copy gains a hash-verified lifecycle
(R2-8); the dead-man surface gains a SessionStart push channel (R2-9) and a per-cause
HALT (R2-10); invariant 7 added. (Process note, owned: this revision shipped without its
CHANGELOG/debate.md blocks — repaired at round 3 with catch-up entries.)

**Round 3 revision (2026-07-17):** all 17 round-3 gaps addressed additively; invariant 8
added (guards specified over ALL exit paths, authorship surfaces, and execution modes).
Design deltas of record: rung 0 invokes the SAME wrapper via `--manual` (R3-2 — the
R2-16b out-of-ledger accept dissolves); the dontAsk read-only Bash carve-out is
deny-enumerated (R3-14) and the `git log --output` write gadget is pinned/denied/
hook-matched shut (R3-15 — both leaf-reverified by blue); snapshots run at every exit
with chained compares (R3-7) and the window log survives dead runs (R3-4); the
red-memory surface is provenance-tagged (R3-3); the infra-class bypass is
wrapper-event-log-keyed (R3-5); the dead-signature normalization is specified (R3-11)
and keys per-signature doctor counts (R3-6); `graduation-queued` gains a queued-stale
backstop (R3-13); the mid-drive `--json-schema` leg is demoted to verify-at-build with
a named fallback (R3-1); §0's enumeration and the tokenizer/telemetry citations are
completed (R3-10/R3-16/R3-17); the wrapper root-of-trust residual is owned (R3-9).
(The "carve-out is deny-enumerated" delta above is SUPERSEDED round 4 — see below: the
enumeration was never exhaustive and the git member escaped; the round-4 structural close
is a bare `Bash` deny.)

**Round 4 revision (2026-07-17):** all 16 round-4 gaps addressed additively. Cross-cutting
shape (the lead's round-4 direction): red's round-4 catches were "the round-3 repair closed
the path it named but not the SIBLING of that path" — answered where possible by STRUCTURAL
closes derived from invariants 6/8, not another lap of neighbor-by-neighbor enumeration.
Design deltas of record: the **whole dontAsk read-only-Bash carve-out class is closed at the
tool boundary** by a bare `Bash` deny in the sleeper profile (R4-3 — enumeration was
non-exhaustive; the session never invokes Bash), with the git channel additionally inverted
to a read-ALLOWLIST at the hook (R4-2 — `git format-patch -o` and the
archive/bundle/config/gc writer family were siblings the `--output` denylist missed;
`-o`/`-O` short forms and the writer family added to belt denies too, R4-2/R4-15); the
enumerated per-command denies are RETAINED as belt for rebuilt rungs. `/self-improve` becomes
a thin `disable-model-invocation` trampoline with its payload moved out of `commands/`, and
§3.4's containment polarity is corrected (markerless out-of-contract dirs are NON-sleeper and
CAN corroborate — R4-1). Red-memory writes are DECLARED DENIED by design under the sleeper
profile (nightly seats do not learn; the R3-3 machinery is belt for drift, tagging at
window-added granularity — R4-4). The unobserved-exit window sweep is bounded by the next
wrapper START and confined to sleeper date-key naming (R4-6). The status re-surface ROOT
INVARIANT is stated once, ending the per-status patch chain (R4-9). The R4-11 gadget
attribution is re-scoped honest (both probes ran under auto mode; the isolating probe
deferred to build). Plus the cheap fixes: qmd-degrade surfacing (R4-7), FOUR-code-artifact
count + host plugin (R4-8), cap/HALT arithmetic recompute (R4-10), est_complexity source
(R4-12), gate-survival provenance row (R4-13), dead-man flag custody (R4-14), CHANGELOG R3-8
bullet (R4-16).

**Round 5 revision (2026-07-17):** all 10 round-5 gaps addressed additively. Red's round-5
thread — "the round-4 structural closes were derived over the TOP-LEVEL session, not the
surfaces where nightly work executes" — is answered with the lead's own remedy shape: state
the invariant once over the TRUE surface. Design deltas of record: the seat-population horn
is PICKED — the FEOV seat agents are INSIDE the sleeper boundary as a design requirement,
with layer-1 `--settings` inheritance a named build probe (OQ23(e)) and the sleeper-guard
hook generalized to deny the WHOLE Bash channel for sleeper-marked runs per invariant 6, so
seat-surface closure never rests on the unprobed inheritance assumption alone (R5-1; the
nightly seat-Bash capability cost is owned in §2.2 step 4 / §5.2, the four dead `Bash(git
…)` allow rules are re-labeled, and the refuted "auto-approves read-only git regardless"
comment is corrected at both sites). The **corroboration-pool ROOT INVARIANT** is stated
once, mirroring R4-9: the pool is POSITIVELY defined — a dir corroborates only with
affirmative non-sleeper provenance (pre-epoch date or interactive-origin marker stamped at
creation); anything harvest cannot positively attribute is QUARANTINED (counts-for-nothing,
fail-closed), dissolving the three round-4 per-surface residuals and promoting OQ24's
quarantine from deferred to BUILT in generalized form (R5-3). The §3.4 rung-0 cell and the
Phase-4 acceptance test are rewritten to the R4-1 trampoline shape — the test is now
two-legged, with trampoline INERTNESS the verifiable property (R5-2). The §1.5 name-key
doctrine is qualified (name-keying confines retroactive-uncertain sweeps, never assigns
origin) and the wrapper logs every sub-run dir PATH at creation so confinement matches
recorded paths after a hard-kill (R5-4). The `-o` belt is declared KNOWN-incomplete on
attached short forms — the hook allowlist is the close of record, with the attached form
added to OQ18(c)'s probe matrix, not another belt lap (R5-5). The `expected-memory-deny`
bucket gains an actor/shape predicate — any memory-surface denial outside the red-merge
seat's own bounded write shape surfaces as `unexpected-memory-write-attempt`, persisting
like TAMPER (R5-6). Sleeper-service's hooks.json wraps its commands in the bootstrap guard
and the flag-missing anomaly is gated on a prior successful run (R5-7); the §0 tree finally
draws `hooks/hooks.json` + the SessionStart executable (R5-8); status changes carry a DATE
the re-surface arms key on, and the `regression` token's domain is pinned docket-only
(R5-9); the est_complexity note gets a literal token convention, currently unpopulated at
the pin (R5-10).

---

## 0. Design summary (the implementable shape)

```
plugins/sleeper-service/
├── .claude-plugin/plugin.json            # requires frank-exchange-of-views
├── skills/continuous-learning/SKILL.md   # promotion ladder: insight → MEMORY → idea stub →
│                                         #   graduated research → rules/skills (HUMAN-GATED);
│                                         #   expand-existing-before-append discipline
├── commands/
│   ├── self-improve.md                   # THIN TRAMPOLINE (round 4, R4-1): instructs the
│   │                                     #   human to invoke `node <operator-owned>/
│   │                                     #   sleeper-wrapper.mjs --manual`; the full loop
│   │                                     #   PAYLOAD lives in the wrapper's phase-1 prompt
│   │                                     #   (skills/continuous-learning), NOT in commands/.
│   │                                     #   frontmatter: disable-model-invocation: true —
│   │                                     #   scheduled/model fires reach Claude as inert text,
│   │                                     #   matching graduate.md (no instructional-gate-only
│   │                                     #   entry point; §4.1's 0/175 evidence)
│   └── graduate.md                       # idea → full FEOV debate → projects/ promotion PR.
│                                         #   frontmatter: disable-model-invocation: true
│                                         #   [minority: lane-1/adversarial — the specific
│                                         #   frontmatter mechanism; interactive-only doctrine
│                                         #   is convergent across all lanes]
├── scripts/
│   ├── harvest.mjs                       # mechanical signal harvester (Node, cost-audit.mjs
│   │                                     #   class; zero tokens, simulator-testable); runs
│   │                                     #   WRAPPER-SIDE, pre-session (R1-16) — the model
│   │                                     #   session never invokes node; wrapper calls it
│   │                                     #   by absolute path
│   ├── sleeper-wrapper.mjs               # scheduler entry point, runs OUTSIDE the model
│   │                                     #   session: PHASE-DRIVEN session driver over
│   │                                     #   --input-format stream-json (R2-1/R2-3) —
│   │                                     #   preflight, harvest staging, ledger
│   │                                     #   (operator-owned dir — R1-19), canary drive +
│   │                                     #   fired-record check (R1-28/R2-1/R2-2),
│   │                                     #   plugin-copy hash verify (R2-8), start/end
│   │                                     #   every-exit snapshots + chained auto-compare
│   │                                     #   (R1-26/R2-7/R3-7), run-window log (R3-4),
│   │                                     #   resume cap k=3 + per-cause HALT (R1-29/R2-10),
│   │                                     #   dead-man record (R1-29/R2-9)
│   └── sleeper-guard                     # PreToolUse write-fence: deny writes outside
│                                         #   research/ + ideas/ for sleeper runs; also denies
│                                         #   the whole Bash channel for sleeper-marked runs
│                                         #   (R5-1 belt — invariant 6); fence LIVENESS proven
│                                         #   per run by the step-0 canary's hook-authored
│                                         #   fired-record (R2-2) — marker loss or hook
│                                         #   silence aborts the run instead of failing open
│                                         #   (R1-28); deny-set baked into the binary, not
│                                         #   read from editable config; REGISTERED in
│                                         #   hooks/hooks.json below (R5-8)
├── hooks/
│   ├── hooks.json                        # registers the sleeper-guard PreToolUse fence AND
│   │                                     #   the SessionStart staleness hook (R5-8 — the tree
│   │                                     #   previously shipped neither registration: three of
│   │                                     #   four code artifacts and no hook wiring); every
│   │                                     #   command wrapped in the bootstrap guard (R5-7)
│   └── sc-sleeper-staleness              # the SessionStart staleness-warning executable —
│                                         #   the FOURTH code artifact (R4-8), drawn in the
│                                         #   tree only now (R5-8: the enumeration was
│                                         #   repaired three times while the tree never was)
└── docs/scheduling.md                    # the ladder: manual → OS scheduler + claude -p →
                                          #   Desktop scheduled task → cloud Routine → GitHub
                                          #   Actions; preflight + ceilings for every
                                          #   unattended rung
```

New code surface is deliberately small — **exactly FOUR new code artifacts (count corrected
round 4, R4-8 — round 3's "THREE" headline never reconciled the SessionStart hook R3-10
enumerated as "a new executable + hooks.json registration"):** harvest.mjs,
the sleeper PreToolUse guard, the scheduler wrapper (round 0 said TWO — the wrapper
was already implicit in §3.4's rung-1 row, and round 1 promotes it to a named artifact
because five gate-side controls now live in it: preflight, ledger, canary, snapshot,
resume cap), and the **SessionStart staleness-warning hook** (a small executable + its
hooks.json entry, minted round 2 R2-9; **host plugin: sleeper-service** — it must fire in
interactive sessions on this box independent of the sleeper scheduler, so it ships with
the sleeper plugin's own hooks.json, R4-8 — **and that hooks.json wraps EVERY command in
the same bootstrap guard prosthetic-conscience's does (R5-7): the sleeper SessionStart hook
fires in every interactive session, so an unguarded sleeper hooks.json during a
plugin-cache empty-bin window would crash-storm all interactive work, a surface §6 row 9's
nightly-scoped bound never contemplated; row 9's bound now re-points at THIS file, and the
empty-bin acceptance check (OQ10) covers it**); plus two command prompts, a scheduling doc, **the continuous-learning skill
file, and the plugin manifest** (round-2 totality fix, R2-19: the two latter entries are
printed in the tree above but were absent from round 1's enumeration, which also made
"everything else reuses shipped machinery" literally false for them — they are new PROSE
artifacts, not code, and are now enumerated); **round-3 totality fix (R3-10 — the round-2
sweep was verified against the round-1 tree, so the round-2 additions were themselves the
omitted specimens):** the enumeration further includes the SessionStart staleness-warning
hook and its hooks.json entry (minted round 2, R2-9 — now COUNTED as the fourth code
artifact and host-plugin-named above, R4-8), the doctor-line delta — which is a
**CROSS-PLUGIN change**: the dead-man/TAMPER/HALT lines land in prosthetic-conscience's
`/doctor`, an edit to that plugin, not to sleeper-service — **a second cross-plugin edit
added round 5 (R5-3): the interactive-origin stamp is one line in FEOV's
setup-research-run.mjs (every run dir it creates gets `inputs/.run-origin` at creation),
an edit to frank-exchange-of-views, not to sleeper-service** — and the two operator-owned
config artifacts (`sleeper-permissions.json`, `sleeper-mcp.json`), which are new
operator-deployed surface even though they never enter the repo; everything else
reuses shipped machinery (FEOV smoke mode, setup/capture scripts, permission engine, existing
hooks) [minority: lane-3/local-probe — the explicit artifact count; the reuse posture is
convergent].

**Loop invariants (each argued from evidence in §§1–5):**

1. **Consume durable artifacts, never introspection** — friction.md harvests, cost.md,
   board-telemetry.jsonl, run-report open questions, ideas/backlog.md + doubts.md, red's
   gap-pattern memory (mirrored). If a wanted signal isn't in a durable artifact, fix
   capture, not recall.
2. **Mechanics cheap, judgment full-strength** — a deterministic script ranks; the ONE picked
   item gets real research; the stub's job is provenance + alternatives, not conclusions.
3. **Headless is viable — proven by live probe, not inference [minority: lane-3/local-probe
   for the probe; doc-level viability is convergent] — but must be verified per run, not
   assumed**: the scheduler wrapper asserts plugin load from `system/init` before trusting
   the run.
4. **Consent gates are structural and layered; at least one layer sits outside the model's
   write reach entirely.** Permission rules alone are empirically insufficient (§4).
5. **Every unattended run carries a hard per-run budget (`--max-budget-usd`), a turn cap, a
   month-to-date ledger preflight (ledger operator-owned, outside the write surface —
   §5.1, R1-19), and leaves a resumable honest partial on any abort (resume attempts
   capped at k=3 — R1-29).**
6. **The design invariant, stated once:** *the loop's write surface and the suite's behavior
   surface are disjoint by permission-engine enforcement, and the mapping is
   allowlist-defined, never denylist-defined.* The loop proposes in its own space; every
   promotion crosses a boundary only a human action can cross, and the boundary's enforcement
   lives where the loop cannot write.
7. **The wrapper-gate invariant (added round 2 — the lead's cross-cutting direction, from
   the R2-1/R2-2/R2-7/R2-8 shape):** *every wrapper-hosted gate emits a liveness/outcome
   record that the wrapper itself CHECKS in-run and that a human SURFACE reports* — a
   control that records or attempts without verifying-and-surfacing its own outcome is
   telemetry, not a gate. Derived instances: the canary checks the hook's fired-record and
   aborts on absence (R2-1/R2-2); the every-exit snapshot compare raises a flag and fails
   the next preflight closed on delta or on a missing chain link (R2-7/R3-7); the preflight verifies the executing plugin
   copy's hash against the operator-approved value and aborts on mismatch (R2-8); the
   dead-man/HALT records carry reasons a doctor line and the session-start warning print
   (R2-9/R2-10/R2-18). The checked-record discipline is the design's answer to
   round 2's board, stated once so future gates inherit it instead of re-learning it
   gate-by-gate.
8. **The completeness invariant (added round 3 — the lead's cross-cutting direction, from
   the R3-2/R3-3/R3-4/R3-7/R3-11 shape: repairs that held for the scheduled/completed
   happy path but not the default mode, the abort path, or the sibling surface):** *every
   guard is specified over ALL exit paths, ALL authorship surfaces, and ALL execution
   modes — the default/manual mode included.* A guard that holds only for the scheduled,
   completed, happy path is a partial guard wearing a total claim. Derived instances:
   rung-0 manual runs invoke the SAME wrapper via `--manual`, so ledger, origin markers,
   canary, and staging exist in the default mode (R3-2); snapshots are taken at EVERY
   wrapper-observed exit path, not only step 7, and the cross-run compare chains to the
   last recorded snapshot of any kind (R3-7); the run-window START is logged at step 0,
   so a run that dies resumeless still leaves its window on record (R3-4); causal
   provenance tagging covers the red-memory mirror surface, not only run dirs (R3-3);
   the dead-signature normalization is specified against the design's own
   fresh-paths-per-night convention (R3-11).

---

## 1. H1 — What the loop consumes: artifact mining, not introspection (SUPPORTED)

### 1.1 The case against artifact-mining (disconfirming pass, hunted first)

- **Telemetry noise / alert fatigue.** Operations practice reports that most raw telemetry
  alerts are never acted on and teams desensitize; vendor analyses put the acted-upon
  fraction under one in five (figure seen at search-digest level only — NOT leaf-verified;
  carried as qualitative direction, not a number).[^AlertFatigue]
  [minority: lane-1/adversarial]
- **Proposal volume, not proposal quality, is the observed failure mode of daily automation.**
  The Dependabot literature is the closest long-baseline field study of a daily autonomous
  improvement loop feeding a human gate: developers respond to volume by configuring toward
  fewer notifications, and 11.3% of studied projects deprecated the tool outright; the 2025
  follow-up frames the core problem as alert fatigue.[^Dependabot][^DependabotFatigue]
  Consequence: the loop needs a **work-in-progress cap** (one stub per run; dedupe against
  open stubs before minting) far more than it needs throughput.
  [minority: lane-2/primary-literature — the Dependabot evidence; the 1-stub cap itself is
  convergent across all lanes]
- **Goodhart / metric gaming.** When a measure becomes the optimization target it stops
  measuring; the reward-hacking literature documents agents satisfying the literal metric
  while defeating the intent (CoastRunners lagoon; proxy/true-objective divergence under
  continued optimization).[^Goodhart] A loop that scores itself on "friction items closed"
  or "proposals emitted" can game those counters. [minority: lane-1/adversarial]
- **Signal that evaporates.** The corpus itself proves some in-run signal is session-local:
  `log()` is operator-console-ephemeral (settled by direct check, run 4), the friction array
  was script-local until the terminal return (run 3, R5-6), and raw trajectories "currently
  evaporate with the session" (backlog item 10).[^FrictionRun4][^Backlog]

### 1.2 Why artifact-mining survives the attack

- **The friction harvests are pre-curated, not raw.** Each entry is a seat's judgment-shaped
  complaint (capability named + what-I-would-have-done), aggregated by the lead — run 3 is
  17 attributed entries, the efficiency run's ~30–39; both already read as ranked backlogs:
  PDF extraction was reported by red, blue, AND judge across all four rounds of one run
  (backlog 27c: "TOP TOOL GAP ... across all 4 rounds"), and recurred in red's merge-seat
  friction across two consecutive runs (backlog 31h) — two sources, stated separately
  (round-1 correction R1-2: the round-0 sentence fused them into a three-seats ×
  two-runs claim stronger than either source); the Read-cap class recurred at six
  seat classes in one run with counts attached; the write-guard class at five consecutive
  round-seats.[^FrictionRun3][^FrictionRun4][^Backlog] The alert-fatigue failure mode
  belongs to unfiltered machine telemetry; this input is closer to postmortem findings than
  to alerts. No curation stage was needed to rank these — recurrence × severity is literally
  present in the text. The H1 falsifier (telemetry too noisy to rank) is disconfirmed on the
  two harvests we have; re-check once harvests from bounded daily runs (smaller, noisier)
  accumulate.
- **Intrinsic self-correction without external signal is the refuted alternative.** LLMs
  "struggle to self-correct their responses without external feedback, and at times, their
  performance even degrades"; prior claimed gains depended on oracle
  feedback.[^SelfCorrect] Loops that work (Reflexion-class) convert *execution feedback*
  into persistent verbal memory consumed on the next episode[^Reflexion] — structurally what
  friction.md + gap-pattern memory + cost.md already are. Voyager grows an ever-expanding
  skill library through "environment feedback, execution errors, and self-verification" —
  durable, compositional artifacts retrieved later, which "alleviates catastrophic
  forgetting."[^Voyager] An introspective /self-improve ("read my own rules and have
  opinions") is the design the external evidence argues against; the prompt must be "here is
  the harvested artifact evidence for defect X; research how X should evolve," never
  "reflect on your rules and suggest improvements."
- **The strongest self-improving-agent systems all share one architecture** — a cheap,
  mechanical outer loop (enumerate candidates, score against recorded evidence, archive
  lineage) around an expensive, delegated inner engine, with empirical validation before any
  change is admitted: Darwin Gödel Machine (archive of agents; empirical validation over
  proof; "transparent, traceable lineage of every change")[^DGM][^DGMSakana]; SICA
  (archive-driven proposal step; 17–53% gains on a SWE-bench Verified subset)[^SICA]; STOP
  (a ~page-of-code seed improver; intelligence entirely in the delegated call).[^STOP]
  Selection by recorded evidence, improvement by the strongest available reasoner — the same
  division this design makes. [minority: lane-2/primary-literature — the DGM/SICA/STOP
  corpus; the architecture claim itself is convergent]
- **Goodhart is contained by separating signal from objective.** The picker's scores rank
  *attention*, never a success metric the loop reports on itself; the only self-graded
  artifact the loop emits is an idea stub a human triages, and promotion runs the full
  adversarial FEOV pipeline where red's job is exactly to catch gamed claims. The loop
  optimizes nothing end-to-end; it proposes. [minority: lane-1/adversarial]
- **Evaporating signal → add capture (H1c confirmed).** The run-4 efficiency debate already
  ratified the pattern: durable signal goes to a git-tracked sink with named consumers
  (`trajectories/board-telemetry.jsonl`, one JSON line per round), explicitly because
  `log()` persists nowhere.[^EffReport] Sleeper-service inherits the rule: anything the loop
  will consume must be written to a git-tracked run artifact at generation time; if a wanted
  signal isn't tracked, the loop's proposal is "add capture," never "parse ~/.claude"
  (blue-respond-r2 spent a seat-round excavating undocumented session internals to settle
  log() persistence — the anti-pattern to design out).[^FrictionRun4]

**Verdict on H1: SUPPORTED, with the noise caveat absorbed by input selection** (curated
complaint/measurement artifacts, recurrence thresholds, human triage at the stub gate).
Confidence: HIGH on the corpus claims (pinned artifacts read directly); HIGH internally that
two harvests demonstrate sufficiency, MEDIUM that it holds for smaller daily-run harvests;
MEDIUM on the external generalizations.

### 1.3 The input inventory (all durable, all git-tracked at known paths)

| Source | Signal class | Proposal class |
|---|---|---|
| `research/*/friction.md` (in-run appended; survives aborts) | pre-ranked capability/protocol gaps with seat+round recurrence | tooling adoption, protocol amendment |
| `research/*/cost.md` (cost-audit.mjs output) | spend physics; per-seat-round tokens/dollars; spike localization | efficiency levers (mechanics only) |
| `research/*/trajectories/board-telemetry.jsonl` | round-by-round board profile, mass trend, new-mint profile (specified by the ratified plan's PR-A.1; **SHIPPED as of FEOV 0.7.0 — present in this run's own `trajectories/`**, corrected round 3, R3-16: the plan's future tense was faithful to the plan but stale against the executing suite — a Phase-4 builder reads this input as ALREADY AVAILABLE)[^EfficiencyPlan] | debate-engine tuning; termination evidence |
| run records / `journal.jsonl` / `run-record-audit.md`[^ResearchCommand] | lifecycle events, abort states, integrity verdicts | resilience fixes (null-guards, resume) |
| run `report.md` "open questions carried past this run" | declared research debts | research topics |
| `ideas/backlog.md` (25 statused checkbox items across 39 lines, with run provenance — recounted at the pin, R1-1), `ideas/doubts.md` (hypothesis → adjudication lifecycle)[^IdeasCorpus] | existing dispositions — the DEDUPE surface; aged open items; staleness signal | doubt-adjudication topics |
| red's gap-pattern agent memory, mirrored into each run's `inputs/red-gap-patterns.md` at pre-create (PR-C.2; this run's mirror is 1,557 lines / 30+ named patterns — byte-exact recount, R1-30)[^EfficiencyPlan] | recurring blue defect classes with how-to-apply lines | checklist lines; protocol clauses |

**Red-memory provisioning (standing friction, now a design requirement):** four seat classes
in the efficiency run could not read red's gap-pattern memory[^FrictionRun4]; sleeper-service
makes the mirror a *harvest step* — executed WRAPPER-SIDE, before the model session starts
(round-1 change, R1-16: the copy is a write, and write-capable mechanics never run inside
the session's Bash allowlist): the wrapper runs harvest.mjs, which copies the red-auditor
memory dir into the loop's run-directory inputs and stages the scored docket, so the model
session opens on a read-complete input set and the loop and any FEOV run it spawns always
have a readable path.

### 1.4 The consumption pipeline: harvest is a script, ranking is arithmetic, the pick is judgment

Per the script-vs-prose doctrine ("an LLM executing mechanics is an unenforced good-faith
contract"[^ResearchCommand]):

**Stage 1 — mechanical harvest (`scripts/harvest.mjs`, zero tokens).** Parse every run's
`friction.md` + `cost.md` + `board-telemetry.jsonl` + `run-record-audit.md`, plus
`ideas/backlog.md` checkbox state and the red gap-pattern mirror headers, into a normalized
signal docket: `class | occurrences (run, seat) | seats affected | max severity seen |
first/last seen | staleness | open backlog item? | existing disposition | score`. Entries
cluster by **defect class, not exact string** — the corpus already names classes consistently
enough for keyword clustering ("write guard", "Read cap", "PDF", "heredoc"), and the corpus's
own lineage lesson says identity-keyed detectors never fire when each cycle mints fresh ids;
carry a `supersedes` field when a signal continues a known class [minority:
lane-1/adversarial — the explicit `supersedes` field; class-keyed dedupe is convergent
lanes 1+3].[^Backlog] Anything already marked DONE/FIXED in backlog.md is closed signal —
recurrence after a fix is its own high-value class: regression [minority:
lane-1/adversarial].

**Stage 2 — ranking (deterministic, stated in the command file so it is auditable):**
`score = recurrence_across_runs × severity_proxy (seat-classes affected) × staleness_decay`,
with lane-2's additional factor `× (1 / est_complexity)` [minority: lane-2/primary-literature
— the complexity divisor], ties broken toward items with a measured cost attached [minority:
lane-1/adversarial — the tie-break]. **est_complexity has a NAMED source, not a model
guess (round 4, R4-12 — a judgment quantity in the mechanics stage is either a hardcoded
guess undermining "auditable", a model estimate moving judgment into zero-token mechanics,
or silently dropped):** it defaults to **1 (inert — the factor vanishes)** unless the
class's matching `ideas/backlog.md` entry carries a human-recorded complexity note, in which
case harvest parses that note's value. **The note has a stated FORMAT (round 5, R5-10 — a
zero-token parser over free prose either guesses, the exact defect R4-12 closed, or never
matches): a literal `cx:<1-5>` token anywhere in the backlog entry line; anything else —
prose, absence, malformed — is the default 1.** And the honest status of the source at the
pin, stated so the text does not imply activation that cannot occur: red's L1 leaf check
(`git show 7bc501e:ideas/backlog.md`, exhaustive token grep across all 25 items) found NO
structured complexity field on any entry — the divisor is universally inert against the
current corpus, and `cx:` is a forward-looking curation surface the operator populates as
triage happens, not a live input today. So the divisor contributes only where a human
already recorded the judgment; the ranking stays pure arithmetic over parsed inputs and the
picker spends zero judgment-tier tokens. Tunable constants live in the script. The formula is
mechanics; it may be wrong, and that is acceptable because the human sees the full ranked
table in the idea stub's provenance and the picker never spends judgment-tier tokens. Skip
any candidate with an open stub *younger than the staleness window* (default 30 days)
already in ideas/; an older untriaged stub auto-stales — harvest recomputes age from the
stub's dated filename each run and re-enters the class in the docket flagged `stub-stale`
(round-1 mechanism, R1-22: without aging, the skip rule composed with gate inattention
would let every untriaged stub permanently subtract its signal class, and the picker would
descend monotonically into noise — the Dependabot base rate applied to our own gate).
**Round-2 refinement (R2-11): auto-stale applies to UNTRIAGED stubs only.** A stub whose
`status` field the human has set to `graduation-queued` (a triage act — §2.3) is EXEMPT
from the 30-day auto-stale while still deduping its class, because graduation lead time
for a heavy human-present event is realistically weeks — longer than a window tuned
against inattention noise — and without the exemption a good queued stub would stale,
re-enter the docket, be re-researched, and mint a fresh stub whose age resets, cycling
indefinitely. The two timers now measure different things: 30 days bounds human
INATTENTION (untriaged); `graduation-queued` is attention already paid. The window stays
operator-tunable in the script constants against observed graduation cadence.
**Round-3 backstop (R3-13) — the exemption is not forever:** R1-22's governing principle
(no stub may permanently subtract its class without a staleness re-surface) applies to
the queued state too — queue-and-forget under deprioritization or operator turnover is
inattention wearing a triaged label, and an eternal exemption would re-open the
monotonic-blinding failure for exactly the queued subset. A `graduation-queued` stub
older than a cadence-tuned M (default 90 days, operator-tunable beside the 30-day
window) re-surfaces in the docket flagged `queued-stale` for human RE-CONFIRMATION: one
keystroke re-arms the exemption for another M; no response converts the stub to ordinary
`stale` handling. The two backstops measure inattention at two cadences — no status is
timer-free (§2.3 carries the flag value; §6 row 3 re-affirmed).

**The root invariant, stated ONCE (round 4, R4-9 — ending the per-status patch chain
R1-22 → R2-11 → R3-13; the third consecutive per-status fix is the missing-root-invariant
signal):** *every status's dedupe effect has a stated re-surface path — no status subtracts
its signal class from the docket permanently.* This governs the two TERMINAL states the
round-3 enum left timer-free: (a) **`graduated`** — a graduated stub dedupes its class, but
if the class RECURS after graduation (the fix regressed, or the graduation did not cover the
observed occurrence), harvest re-enters it flagged `regression`, mirroring the backlog
DONE/FIXED regression rule (§1.4 stage 1); graduation is not permanent immunity. (b)
**`rejected`** — a rejected stub dedupes its class for a cadence-tuned window (default 90
days, operator-tunable, OR until the class's recurrence rate exceeds its pre-rejection rate,
whichever first), then re-surfaces flagged `rejected-recurring` for one-keystroke
re-confirmation; a single rejection cannot permanently subtract a class that keeps recurring
(the monotonic-blinding failure R1-22 forbids), and a re-confirmed rejection re-arms the
window (the Dependabot-fatigue arm — no re-mint every run). **Both arms of (b) key on the
REJECTION date, which round 5 gives an artifact (R5-9 — stub filenames date the MINT, so a
filename-keyed implementation would measure mint age, not rejection age: a stub rejected at
day 80 would re-surface in 10 days; and an undated rate-arm silently drops — the
printed-invariant≠built-invariant class):** every human status change appends its date to
the status line (`status: rejected 2026-07-17` — §2.3 carries the format); harvest parses
that date and both arms compute from it. A terminal status with NO date is fail-toward-
re-surface: harvest treats it as set at MINT (the earliest plausible reading, so it
re-surfaces sooner, never later) and flags the stub `undated-status` once for the human to
date it — a dateless rejection can shorten its own dedupe window but can never extend it.
Every status now has its re-surface path named: `open`→30-day stub-stale; `graduation-queued`→90-day queued-stale;
`graduated`→regression on recurrence; `rejected`→90-day/rate rejected-recurring;
`stale`/`queued-stale`/`rejected-recurring` are themselves the re-surfaced docket entries.
**Token domains pinned (R5-9 — the round-4 enum left `regression` unplaced and the
`rejected-recurring` setter unstated):** the eight tokens split into two disjoint domains.
STUB-FILE statuses (`open` at mint; `graduation-queued`/`graduated`/`rejected` set by
humans, dated) live in the stub and only humans edit them — the loop never edits existing
ideas/ files (§1.5). DOCKET flags (`stale`, `queued-stale`, `rejected-recurring`,
`regression`, `undated-status`) are computed by harvest each run and exist ONLY in the
staged docket — harvest writes no stub. So `regression` is a docket-flag, full stop: the
graduated stub STAYS `graduated`; the regression docket entry may motivate a NEW stub
(fresh file, fresh date) whose provenance cites the graduated ancestor. `rejected-recurring`'s
setter is harvest (docket-only); the human's re-confirmation act is re-dating the stub's
`rejected` status line, which re-arms the window. §2.3's status enum and §6 row 3 carry
this.

**Stage 3 — the model picks ONE** — a judgment call, but a cheap one: the worst failure is a
suboptimal research topic for one bounded run (self-correcting next day). Risk-accepted at
the bulk tier. Think-around-problem still applies to the pick: the prompt requires the
docket's top 3 be compared, not just top-1 taken [minority: lane-3/local-probe — the
compare-top-3 requirement].

**Falsifier fallback (noise/lag):** `ideas/backlog.md` is the human-curated intermediate
between raw friction and action, and friction→backlog graduation has happened by hand after
every run.[^IdeasCorpus] The loop reads BOTH: raw harvest for recency, backlog state for
curation. If harvest-ranking proves noisy in runs 5–7, the degradation path is "loop proposes
from backlog only" — judgment cost stays zero because the curation judgment was already spent
by the human. [minority: lane-3/local-probe]

**The null alternative, priced honestly (added round 1, R1-14):** the demonstrated by-hand
loop — human harvests friction, graduates items to backlog, picks what to fix — has run
after every run and works; daily automation must argue its margin over it, against the
report's own Dependabot evidence that daily automated proposal cadence is where gate
fatigue lives.[^Dependabot][^IdeasCorpus] What the automated loop buys that the by-hand
loop demonstrably does not do: (a) the bounded research pass and the structured stub —
backlog entries are one-to-three-line dispositions, never provenanced stubs with 3–5
compared alternatives and an acceptance shape; (b) mechanical recurrence × staleness
arithmetic over a corpus already 1,557 lines of gap patterns plus multi-run friction —
by-hand ranking satisfices as the corpus grows; (c) removing the human as single point of
recall for cross-run recurrence. What it costs: expected ~$0.10–0.50/night at bulk-tier
list rates (the ~50k-token smoke shape[^Backlog]; probe P2's whole plugin-command run cost
$0.058[^HeadlessProbe]) under a $2–5/night hard ceiling and a ~$50/mo ledger cap — the
ceiling and cap are anomaly bounds, not expected spend (R2-18; arithmetic reconciled in
§5.2) — plus the fatigue risk at the gate. The priced conclusion: **rung 0 — manual
`/self-improve`, the SAME wrapper code path invoked by hand (`node sleeper-wrapper.mjs
--manual`, decided round 3, R3-2 — §3.4; "same code path" is now true by construction,
not assertion), zero standing cost — is the DEFAULT and may be
terminal**; scheduling is opt-in (port plan resolved decision 6[^PortPlan]) and daily
cadence is a hypothesis the stub-survival ledger tests, not a given. Named revisit
trigger, written into scheduling.md: if fewer than ~1 in 3 stubs receive human triage
within the staleness window, the cadence drops (weekly, or back to rung 0) — the
Dependabot lesson applied to ourselves before the fatigue, not after.

### 1.5 Self-poisoning guard [minority: lane-3/local-probe]

From red's memory-poisoning pattern class: the loop writes idea stubs as NEW files
(`ideas/<date>_<slug>.md`) and NEVER edits `ideas/backlog.md` or its own harvest inputs.
Past run dirs are convention-immutable (the pinning convention), so the harvest reads a
stable corpus. (See §2.2 step 6 for the conflict with lane-2's backlog-append step.)

**Round-1 extension — the guard must cover AUTHORSHIP, not only edits (R1-25).** The
never-edits rule bounds mutation of existing inputs, but the nightly run *creates* a run
dir whose friction/cost artifacts are model-authored and land in the next morning's
harvest corpus — "run-record authored" includes loop-authored, so without more, a friction
class the loop emits nightly would ratchet its own recurrence × staleness score with no
human in the sub-loop. Needs no malice: noisy or injection-tainted entries (§6 row 14)
suffice. The authorship guard, mechanical and wrapper-side, in two decided parts:

**Tagging is by CAUSAL PROVENANCE, never by name (round-2 repair, R2-5 — the round-1
`*_self-improve/` name glob was circumvented by the design's own default control flow: the
nightly run creates TWO run dirs, and the spawned bounded-FEOV pass's topic-slug dir
escaped the glob, landing untagged loop-authored friction that would have both counted at
full recurrence AND supplied the very "non-sleeper corroboration" the cap demands).**
Mechanism: the wrapper stamps an origin marker file (`inputs/.sleeper-origin`, carrying
the parent run id) into EVERY run dir it creates — the `*_self-improve/` dir at preflight
AND the bounded-pass sub-run dir, which the wrapper itself now creates when it runs
setup-research-run.mjs wrapper-side at phase B (§2.2 step 4; the round-2 R2-3 relocation
is what makes stamp-at-creation possible). Belt: at step 7 the wrapper diffs its start/end
porcelain snapshots (already taken — R1-26) and stamps any run dir that appeared during
the run window and lacks a marker; harvest.mjs reads ONLY the marker, never the dir name —
**qualified round 5 (R5-4: the unqualified round-2 sentence stood beside the round-4
confinement clause that decides sweep membership BY name — a doctrine-vs-mechanism
contradiction as printed): name-keying is permitted ONLY to CONFINE a
retroactive-uncertain sweep (narrowing what an uncertain window may auto-tag), never to
ASSIGN origin outside one — origin assignment is always marker or recorded window/path
evidence** — and harvest additionally treats any run dir whose creation timestamp falls
inside a wrapper-logged sleeper run window as sleeper-origin even if markerless. **The window log survives dead
runs (round-3 repair, R3-4 — the round-2 window was keyed on ledger timestamps, which are
written only at step 7, so a run that aborted and exhausted its resumes never logged a
window and both backstops died with it):** the wrapper logs the run-window START in the
operator-owned dir at step 0 (beside the start snapshot, which already lives there) and
the END at whatever exit it observes — step-7 record, abort, or DEAD-mark time — so even
a run that dies resumeless leaves its window on record; and the origin-marker porcelain
sweep runs at DEAD-marking time as well as at step 7, so a model-authored in-session run
dir created on a night that later dies is stamped rather than entering harvest markerless.
**The UNOBSERVED-exit case is owned round 4 (R4-6):** a wrapper hard-kill (SIGKILL, power
loss, OS kill — §6 row 6's own High-likelihood environment) is an exit the wrapper never
observes, leaving a START with no END until DEAD-mark time ~3–4 days later; the naive
multi-day window would then sweep daytime HUMAN-present run dirs into `origin: sleeper`,
capping them at 1 and EXCLUDING them from the corroboration pool — suppressing exactly the
human-present corroboration ordinary classes wait on. Three clauses close it: (i) a window's
END is additionally bounded by the NEXT wrapper START — no window spans invocations, so the
sweep of an unclosed window can never reach past the following night's fire; (ii) a
DEAD-mark-closed window (START with an unobserved exit, closed retroactively) is flagged
`retroactive-uncertain`, and its markerless sweep is CONFINED to dirs bearing the sleeper
date-key naming convention (`research/<date>_self-improve/`) **plus the RECORDED sub-run
paths (R5-4 — "the wrapper's own sub-run slug" was not a knowable convention after the
hard-kill the clause exists for: the slug is model-chosen per night,
pattern-indistinguishable from a human's same-day research dir, and the wrapper that knew
it died with it recorded nowhere durable. Fixed by recording, not by pattern: the wrapper
appends the sub-run dir PATH to the run-window log AT CREATION, beside the START record —
it creates the dir itself at phase B (R2-3), so the path is in hand pre-kill — and
confinement matches recorded paths exactly. The mkdir-to-stamp bound, stated explicitly:
the marker stamp and the path log-append happen in the same wrapper code block immediately
after mkdir, so the unstamped-but-existing window is one synchronous statement wide; a kill
inside that statement leaves a dir the sweep still catches via the logged path or, if the
kill beat the log line too, a dir with NO durable sleeper evidence — which the positive
pool below quarantines rather than counts, fail-closed)** — any OTHER markerless dir in
that window is surfaced for human confirmation rather than auto-tagged sleeper, so a
human's daytime research dir is never silently swept; (iii)
the snapshot-chain leg survives an unobserved exit regardless (the last recorded snapshot is
the START snapshot, so the cross-run compare still has a referent — §4.3 layer 5), and the
resume path re-opens the same run dir under a fresh observed window. The backstop for
harvest reading an unclosed window: treat it as extending to the present ONLY for
date-key-named dirs, never for arbitrary dirs.
Harvest tags every row from a marked dir `origin: sleeper` and **caps sleeper-origin
recurrence contribution at 1 per class** regardless of how many nights repeat it.

**The red-memory surface: agent-memory writes are DENIED by design under the sleeper
profile — the R3-3 premise is corrected round 4 (R4-4).** Red's gap-pattern memory lives
in `.claude/agent-memory/**`, which the sleeper profile FORECLOSES at three layers: the
`Edit(<REPO>/.claude/**)` deny, the research/+ideas/ write fence, and the phase-0 canary
abort. So the round-3 premise — "the nightly bounded-FEOV pass spawns a red-merge seat that
WRITES the shared agent-memory dir under hook coverage" — cannot be true simultaneously
with §4.2's own profile; the two contradicted. The horn picked, and its consequence stated
plainly: **nightly seats do not learn.** A nightly red-merge seat's mandatory
record-new-patterns write FAILS-DENIED every night under the fence — this is EXPECTED
behavior, and the resulting fence-denial is a NORMAL fired-record class the doctor line does
not flag as anomaly (distinct from the canary/TAMPER/HALT classes; the wrapper's
fired-record parser buckets `.claude/agent-memory/**` denials as `origin=nightly-seat,
class=expected-memory-deny` — **predicate narrowed round 5 (R5-6: keying the bucket on
TARGET PATH alone would have filed an injection-driven memory-poisoning ATTEMPT — §6 row
14's Medium-likelihood channel — as expected at any volume and any target under
agent-memory/**, answering red's distinguishability question by fiat; invariant 7's own
test says a control that records without surfacing is telemetry, and the memory surface
steers the adversary's future audits). The bucket admits ONLY the red-merge seat's own
record-new-patterns write shape — seat role = red-merge AND target = the red-auditor
pattern file — bounded at N≤2 denials per run (the expected event is roughly one per run).
Any other actor, any other subdir under agent-memory/**, or volume over the bound files as
a distinct `unexpected-memory-write-attempt` class that the doctor line surfaces and that
PERSISTS until the human clears it, like TAMPER — enforcement was never in question (the
write is denied either way); what this restores is the tamper-attempt-must-be-VISIBLE
doctrine on exactly the surface where suppression was cheapest**). Nightly gap patterns
reach the corpus only via
ordinary-classed friction TEXT (the run's own friction.md, which IS writable under research/
and IS harvested), never via agent-memory authorship. Red's learning is a
HUMAN-PRESENT-run property: graduation and interactive FEOV runs carry the interactive
profile and write agent-memory normally. The R3-3 snapshot-diff machinery is therefore
retained only as **belt for out-of-window and interactive authorship and defense-in-depth
against profile drift** (a rebuilt rung, or a misconfigured profile that dropped the
`.claude` deny): the wrapper's step-0/exit snapshots (R3-7) extend to the red-memory dir
(file list + content hashes); if a pattern file or header nonetheless CHANGES inside a
wrapper-logged sleeper window — which under the corrected design signals profile drift, not
normal operation — the sweep records it and the change is flagged for human review. **Tagging
is at the granularity of what the window ADDED (R4-4 second leg):** pre-existing pattern
content keeps its origin; only content a sleeper window newly APPENDED is tagged `origin:
sleeper` — red's discipline is to EXTEND existing pattern files, so a single nightly append
must not re-classify a pattern carrying months of human-present evidence to sleeper-origin
(the R1-22 monotonic-blinding leg, one surface over). Sleeper-added content is capped at 1
per class and excluded from the non-sleeper corroboration pool. The mirror staged into run
inputs remains a pre-run frozen snapshot (lead-preferred shape), so mid-run memory writes —
which under the corrected design do not occur in-window at all — never feed the same night's
docket.

**The corroboration requirement is DECIDED, and each horn's cost owned (round-2 repair,
R2-6 — the round-1 "requires non-sleeper occurrence — or, failing that, flags
`sleeper-only`" was two materially different mechanisms wearing one sentence: a hard bar
that structurally blinds the loop to failure classes observable only in its own
headless/--bare/MCP path, or a label whose containment reduces to the human noticing a
flag at the design's own Dependabot base rate).** The decision is the severity-gated
bypass (red's proposed shape, lead-endorsed, adopted): **infrastructure-failure classes**
— aborted-run, fail-closed events (ledger-unparse skip, canary abort, plugin-hash
mismatch), hook-crash, DEAD/HALT markers — enter the docket flagged `sleeper-only`
WITHOUT corroboration. **The infrastructure-class tag is assigned SOLELY from the
wrapper's own event log** — the skip/abort/DEAD/HALT/TAMPER records in the operator-owned
dir, mechanical and session-unforgeable (round-3 sentence, R3-5): friction TEXT never
self-classifies into the bypass lane and only ever contributes to ordinary,
corroboration-requiring classes — §6 row 14 concedes the text channel is forgeable, so a
keyword-clustered bypass would have let a loop-authored or injected entry mentioning the
right words ride the no-corroboration lane through the very channel §1.5 was minted to
distrust. The bypass exists for these classes because they are by construction observable only in
sleeper runs, and a hard bar would blind the loop to its most-relevant failures precisely
as the corpus becomes sleeper-dominated; **ordinary capability/protocol classes** require
at least one non-sleeper occurrence before entering the top-3 comparison. Costs owned in
text: for ordinary classes, a sleeper-only phenomenon waits for human-present
corroboration (accepted — those venues overlap human-present runs, and the docket still
lists the class below the fold, visibly); for infrastructure classes, containment does
lean on the human seeing the `sleeper-only` provenance flag — bounded by the 1-stub/run
cap, the mandatory provenance contract, and the doctor/dead-man line, which round 3 makes
a TRUE second channel for every bypass class rather than an overclaimed one (R3-6: TAMPER
and HALT persist until the human clears them, but canary aborts, ledger-unparse skips,
and hook crashes older than the most recent event scrolled off a single-slot "last
reason" line — leaving the stub flag the sole standing channel for any class followed by
a later skip, the modal case since intermittent flake is the modal infra failure per
#76239/#68375). The line now prints PER-SIGNATURE COUNTS since the last human clear —
keyed by the R3-11 dead-signature normalization, fixing R3-11 and R3-6 with one mechanism
— alongside the most recent reason, so a fired-then-superseded infra class stays visible
and "the flag is not the sole channel" holds as a mechanism, not an assertion. §6 row
10's grade tracks this decision.

**The corroboration-pool ROOT INVARIANT, stated ONCE (round 5, R5-3 — ending the
per-surface patch chain R2-6 → R3-3/R3-5 → R4-1/R4-6/R4-13, the same chain shape R4-9
ended for status timers; red's sharpest case forced the promotion: rung-2 Desktop tasks
run locally against the SAME corpus the nightly harvest reads, so under a
default-negative pool their markerless dirs would count as NON-sleeper corroboration
next morning AUTOMATICALLY, guarded only by an instructional adoption requirement —
§4.1's own 0/175 class — with no human act anywhere in the chain):** *membership in the
non-sleeper corroboration pool is POSITIVE, never default — a run dir corroborates an
ordinary class only when harvest can affirmatively attribute it non-sleeper; anything
harvest cannot positively attribute is QUARANTINED: it neither corroborates nor
contributes recurrence (counts-for-nothing, fail-closed), and appears in the docket
flagged `unattributed` for one-keystroke human attribution.* Positive non-sleeper
attribution has exactly two forms: (a) creation date before the **sleeper deployment
EPOCH** — a timestamp recorded once by the schedule-setup step in the operator-owned dir;
every dir in the pinned corpus predates it by construction, so the existing evidence base
needs no retro-tagging; (b) an **interactive-origin marker** (`inputs/.run-origin`),
stamped at creation by setup-research-run.mjs for every run dir it creates — one line in
FEOV's setup script, the cross-plugin edit enumerated in §0. Precedence, because the
sleeper session can write inside `research/**` and could forge an interactive marker:
sleeper-positive evidence (origin marker, wrapper-logged window or recorded sub-run path)
DOMINATES an interactive marker; a marker CONFLICT (interactive marker inside a
wrapper-logged sleeper window) is quarantined AND surfaced as anomaly. This one predicate
dissolves the three round-4 per-surface residuals: an R4-1 paste-run dir, an R4-6
unobserved-window dir that beat the stamp, and an R4-13 rung-2/3/4 markerless dir all
land post-epoch without positive non-sleeper provenance ⇒ quarantined, never counted as
corroboration — and OQ24's quarantine is thereby PROMOTED from deferred to BUILT, in
this generalized positive-pool form rather than the loop-shaped-dir heuristic (which
would have paid schema-matching false positives on ordinary research dirs; here an
ordinary human dir pays one mechanical marker stamp at setup, and an ad-hoc human dir
created outside the setup script is quarantined-and-surfaced, never silently
misclassified in either direction). The residual named honestly: quarantine adds one
human keystroke per unattributed dir — bounded by the fact that the setup script is the
documented path for creating run dirs, so the steady-state unattributed rate is
paste-runs and hand-made dirs only.

The round-0 sentence "a bad stub cannot amplify its own
future score" is thereby made true by mechanism rather than asserted; §6 row 10 is
re-graded accordingly, and §4.3's "the loop eats its own complaints the next morning" now
carries this cap explicitly.

---

## 2. H2 — /self-improve and /graduate mechanics (SUPPORTED)

### 2.1 Disconfirming pass

- **Are cheap stubs worthless?** In a 100+-expert blind study, LLM-generated research ideas
  were judged *more novel* but *less feasible* than expert ideas, with weak idea diversity
  and unreliable self-assessment; human re-ranking improved outcomes.[^IdeaStudy] An
  unattended loop's raw ideas are exactly the artifact you should NOT trust unreviewed —
  which argues for the stub-then-human-gate design, and *against* both extremes
  (auto-promotion, and demanding the nightly loop produce graduation-grade research).
  [minority: lane-1/adversarial]
- **Does bounded research produce stubs too shallow to survive red later?** (H2's falsifier.)
  Partially defused by lowering what the stub must be: a *well-provenanced pointer with
  alternatives*, not a conclusion. Graduation re-runs full FEOV from scratch; a shallow stub
  costs a wasted graduation candidate, not a corrupted rule. If smoke-grade stubs
  consistently die under red at graduation, the correction is the frontier's own alternative
  — run **less often at full strength** (e.g. weekly full-lane) rather than daily at smoke
  strength. This is measurable, not a design fork: the ledger records stub→graduation
  survival; if <~half of graduated stubs survive round 1 of their promotion debate, escalate
  the bounded mode's scale and drop the cadence. The revisit trigger is written into the
  scheduling doc, not left to memory (same discipline as the efficiency plan's named revisit
  triggers[^EfficiencyPlan]).
- **The bounded mode yields an honest artifact — precedent exists.** The smoke run produced a
  single-round UNVERIFIED verdict with template friction surfaced instead of silently
  degraded.[^SmokeRecord] An expected-UNVERIFIED stamp on a stub is correct, not a failure.
  [minority: lane-3/local-probe]

### 2.2 /self-improve (the daily driver)

```
/self-improve [--budget-usd N] [--model M] [--dry-run]

0. PREFLIGHT (wrapper, OUTSIDE the model session — R1-16/R1-19/R1-26/R1-28/R1-29,
   restructured round 2): plugins loaded (abort loudly if FEOV/sleeper-service absent —
   §3.3); executing plugin-copy hash verified against the operator-approved value —
   FAIL-CLOSED on mismatch (R2-8); qmd daemon /health (degrade note if down);
   month-to-date ledger vs monthly cap (§5 — ledger in the operator-owned dir;
   FAIL-CLOSED on an unparseable ledger: skip with reason, don't guess); HALT marker
   check (R2-10: refuse to launch until the human clears a per-cause HALT); today's
   run-dir idempotency check — completeness derived from the WRAPPER'S OWN step-7 ledger
   record in the operator-owned dir, never from loop-writable run-dir contents (R2-12)
   (recorded-complete → exit 0; started-but-unrecorded → resume, capped at k=3 attempts,
   then DEAD marker written BESIDE THE LEDGER in the operator-owned dir, dated, with
   normalized abort-reason signature + fresh dir next fire — R2-10/R2-12); harvest.mjs
   staging (docket + red-memory mirror) — a harvest FAILURE at staging is itself a
   fail-closed skip (R3-1): nonzero exit or an unparseable docket aborts the run with a
   dated skip reason per invariant 7, never a degraded guess; a down qmd daemon degrades
   rather than aborts, and the degrade note is written INTO the staged docket header AND
   the ledger record, so the stub's §2.3 confidence field and the doctor line are its
   named readers (R3-1); start snapshot (git porcelain + guardrail hashes + red-memory
   dir hashes (R3-3) → operator-owned log; run-window START logged beside it (R3-4);
   compared at every wrapper-observed exit — R2-7/R3-7). The SAME wrapper serves rung 0:
   `--manual` runs the identical phase sequence with console-visible output and an
   `origin: manual` ledger annotation (R3-2 — §3.4).
   THEN the wrapper OPENS the model session as a PHASE-DRIVEN stream-json drive
   (`--input-format stream-json --output-format stream-json`, flag pair verified in
   `claude --help` on the pinned CLI 2.1.212 — R2-1; this reconciles round 1's
   contradictory anchors: the canary ATTEMPT executes in-session, but the ACTOR,
   OBSERVER, and ABORT all belong to the wrapper):
   PHASE 0 — DENIAL CANARY: the wrapper sends a canary-only first message ("attempt one
   Edit to <REPO>/plans/.canary-<nonce>"), then parses the event stream and the
   hook fired-record (§4.3 layer 2, R2-2) BEFORE any real work is prompted. Unless the
   write was DENIED **and** the sleeper-guard's own fired-record contains the nonce with
   decision=deny (proving the HOOK layer specifically is live, not merely layer 1 —
   marker loss means no fired-record, which ABORTS even though the permission profile
   still denies the write), the wrapper kills the session and aborts the run. A post-hoc
   envelope check is refuted by construction (the run would already have executed); an
   instructional in-prompt abort is refuted by the report's own §4.1 0/175 evidence —
   hence the two-phase drive.
1. ENUMERATE (~zero tokens; PHASE 1 — the wrapper sends the real prompt only after the
   canary passes): read the wrapper-staged harvest docket (the session never
   invokes node — R1-16) + inventory of the suite's own surface (rules/skills/agents/
   commands across the three plugins — from the marketplace manifests, not hardcoded
   [minority: lane-3/local-probe — manifest-driven inventory]) +
   ideas/ + open research questions. Read-only on everything outside research/ + ideas/.
2. SCORE: read the wrapper-staged scored table (the §1.4 formula ran wrapper-side inside
   harvest.mjs at step 0 — R2-4: round 1's "script" label was an R1-16 propagation miss;
   no in-session script exists); cheap-tier language only if a tie-break needs prose. The
   full scored table is already logged to the run dir, so the pick is auditable.
3. PICK ONE (model, bulk tier). Exactly one. Compare docket top-3; state why in one
   paragraph. The pick returns to the wrapper because step 4's staging needs it
   wrapper-side — channel DEMOTED to verify-at-build round 3 (R3-1): `--json-schema`
   structured output is documented only for `--output-format json` FINAL results, and
   nothing documents a schema-bound MID-drive phase result under stream-json, so the
   round-2 "the phase drive's `--json-schema` leg" was asserted composition, not
   documented composition. The build probes it (OQ22); the named fallback is the design
   of record until then: the phase-3 prompt requires the pick in a fenced ```pick``` block
   (fixed keys), parsed wrapper-side by deterministic grammar; a message yielding no
   well-formed pick from either channel is a fail-closed skip with reason (invariant 7).
4. RESEARCH "how should X evolve?" (delegated, bounded — EXECUTION LOCUS determined by
   probe round 2, R2-3, from the shipped command at the pin[^ResearchCommand]: FEOV's
   machinery is MIXED — setup-research-run.mjs and capture-research-run.mjs are
   session-Bash `node` invocations in the shipped flow, but the debate engine itself runs
   via the harness's **Workflow tool** (`scriptPath` = debate.js), which is NOT a Bash
   call. Under the sleeper profile the shipped flow as-is would stall on the two Bash
   node calls, so the sleeper drive re-hosts them): PHASE B — the WRAPPER runs
   setup-research-run.mjs wrapper-side (its own pinned argv; the model-chosen topic is
   sanitized to a slug charset before it touches an argument; write targets are the
   script's own fixed code under research/ — and wrapper creation is what lets R2-5's
   origin marker be stamped at creation), then prompts the session to invoke the
   **Workflow tool** on debate.js at smoke scale (1 lane, 1 round, 1 citation pass, bulk
   model — the measured ~50k-token smoke shape[^Backlog][^ResearchCommand]) with the
   harvested evidence staged as pinned inputs and the research-protocol skill preloaded —
   disconfirming budget and semantic footnotes intact. capture-research-run.mjs runs
   wrapper-side at step 7. The Workflow channel's own coverage is argued at §4.3 layer 4
   (re-scoped round 2). **The nightly seat population runs BASH-FREE, and that cost is
   owned here where the capability lives (round 5, R5-1 — the seats sit inside the
   sleeper boundary, so the bare `Bash` deny + the hook's Bash-channel deny bind them):
   the smoke-scale lanes and the citation pass work through Read/Grep/Glob, qmd, and
   WebFetch(arxiv) — the research protocol itself forbids shell file access — and repo
   state (git status/diff/log) is read wrapper-side and STAGED into the run inputs, so
   no seat needs git argv. What is genuinely lost nightly: leaf probes that require
   executing a command (the R3-15/R4-2/R5-5 gadget-reproduction class red's audit seats
   used in THIS run). Priced and accepted: a stub's research pass produces pointers, not
   verdicts; command-execution leaf probes are re-derived at human-present graduation
   where the full profile is live, and a seat that wanted Bash says so in friction —
   which harvest reads as an ordinary class, so the want is visible, not silently
   starved.** Verdict will honestly be UNVERIFIED — correct for a stub.
   Nightly retrieval surface (R1-17/R1-18): qmd corpus + WebFetch(domain:arxiv.org)
   only — WebSearch is NOT in the default nightly profile (human-present graduation runs
   keep it); every web-derived claim is labeled untrusted-origin in the stub (§2.3). The
   nightly profile's strict qmd-only MCP config also means NO pdf-reader/arxiv-latex
   reach (R2-17 — the trade is stated in §3.3 and the stub contract carries the
   confidence consequence).
5. EMIT the idea stub → ideas/<date>_<slug>.md (append-only; never rewrites existing ideas
   or doubts). Contract in §2.3.
6. LINK — CONFLICT CARRIED: lane-2 specifies "append one line to ideas/backlog.md linking
   the stub"; lane-3's self-poisoning guard (§1.5) specifies the loop NEVER edits
   ideas/backlog.md. Both preserved. Proposed reconciliation (synthesis, for red/lead):
   keep the write surface stub-files-only and give harvest.mjs a generated, loop-owned
   index (e.g. ideas/stubs-index.md, machine-rebuilt each run from the stub files) so
   discoverability survives without the loop writing to the human-curated ranking input.
7. RECORD (wrapper, OUTSIDE the model session — R1-19): run capture-research-run.mjs
   wrapper-side (R2-3); append cost JSON + stub path to
   the sleeper ledger in the operator-owned dir (the session has no write path to it);
   end snapshot AND COMPARE (R2-7; completed round 3, R3-7 — the compare ran ONLY at
   step 7 in round 2, so a run dying earlier (deterministic abort, budget/turn kill,
   canary abort, wrapper crash handler, k-th failed resume) took no end snapshot and the
   next run's compare had no referent; aborts are model-influenceable, so that was a
   write-then-force-abort hole in an acceptance leg): the wrapper snapshots at EVERY
   exit path it observes — it outlives the session by construction, so an abort is still
   a wrapper-observed exit, and DEAD-marking time is the exit path for resume
   exhaustion — and the cross-run compare is end-vs-start plus
   start-vs-THE-LAST-RECORDED-SNAPSHOT-OF-ANY-KIND from the prior run, with a MISSING
   chain link failing the next preflight closed per invariant 7. On any mismatch, write
   a dated TAMPER flag beside the ledger, surface it on
   the doctor line, and fail the NEXT preflight closed until the human clears it (the
   window now has a watchman: the wrapper is the comparer, the doctor line + blocked
   preflight are the human surface — invariant 7); origin-marker sweep of the porcelain
   delta (R2-5; also run at DEAD-marking time — R3-4); update last-successful-run +
   skip/abort reason + per-signature counts for the dead-man surface
   (R1-29/R2-18/R3-6); exit with the JSON envelope. [minority: lane-3/local-probe — the
   explicit RECORD step]
```

One item per run is a cost ceiling and a quality floor (the Reflexion analogy: one
reflection per episode, persisted).[^Reflexion]

### 2.3 The idea-stub contract (what makes graduation auditable later)

Schema-checkable by the quality-gate hook [minority: lane-3/local-probe — hook
schema-checking]:

```
# idea: <title>
provenance:        <harvested signal rows that motivated it — friction lines / backlog ids /
                    telemetry rounds, with pins to artifact lines>
alternatives:      <3–5 genuinely distinct (think-around-problem applied to itself), each
                    with a one-line case and cost class>
proposal:          <the favored alternative and why it beat the others>
acceptance-shape:  <what evidence would graduate it; what a landed fix must demonstrably do;
                    what would kill it>
research:          <run-dir of the bounded pass; verdict stamp — expected UNVERIFIED>
confidence:        <self-graded + declared unknowns (the blue pre-flight discipline).
                    Standing nightly caveat (R2-17): the sleeper MCP profile is strict
                    qmd-only, so pdf-reader/arxiv-latex are unreachable and the
                    research-protocol PDF MUST-try clause is unsatisfiable nightly —
                    PDF-capped claims are graded "unable to corroborate (no PDF tooling
                    under the nightly profile)" honestly, and re-verified at graduation
                    where the full MCP surface is present. ALSO (R4-7 — the reader
                    obligation, not just the writer's note): if the staged docket header
                    carries a qmd-degrade note (daemon down at preflight → Grep/Read
                    fallback, §2.2 step 0), the stub's confidence field MUST label affected
                    claims "recall-degraded (qmd down; lexical Grep/Read only this run)" so
                    chronic daemon-start failure surfaces in the artifact a human reads,
                    never silently>
est-cost-to-graduate: <keeper-run + build estimate> [minority: lane-3/local-probe]
graduation_trigger: <recurrence/severity condition under which /graduate is worth a human's
                    time> [minority: lane-1/adversarial]
status:            <STUB-FILE statuses, human-edited and DATED (R5-9): open (at mint) |
                    graduation-queued | graduated | rejected — every human status change
                    appends its date to this line (`status: rejected 2026-07-17`); harvest
                    parses the date and every re-surface arm keys on it (a dateless
                    terminal status is read as set-at-MINT — fail-toward-re-surface — and
                    flagged `undated-status` once). DOCKET flags, harvest-computed, never
                    written to the stub (R5-9 domain pin): stale | queued-stale |
                    rejected-recurring | regression | undated-status | unattributed.
                    Mechanics: harvest auto-stales UNTRIAGED stubs at 30 days from the
                    dated filename (R1-22/R2-11); graduation-queued exempts from the
                    30-day window but re-surfaces queued-stale after M days (default 90)
                    from its status date (R3-13). ROOT INVARIANT (R4-9, §1.4): every
                    status has a stated re-surface path — graduated stays `graduated` in
                    the stub while its class re-enters the DOCKET flagged `regression` on
                    recurrence (a new stub may be minted citing the ancestor); rejected
                    dedupes for a 90-day/rate window from its status date then re-surfaces
                    `rejected-recurring` (harvest-set, docket-only) for one-keystroke
                    re-confirm = re-dating the stub's rejected line; NO status is
                    timer-free>
origin-labels:     <every claim tagged corpus-derived or web-derived(untrusted-origin) —
                    R1-18; web-derived claims never enter ranking inputs>
```

### 2.4 /graduate (the promotion pipeline — human at every step; NEVER scheduled)

1. **Human invokes** `/graduate <idea-stub>` — the loop never self-graduates. The command
   carries `disable-model-invocation: true`: the harness withholds such skills from
   scheduled/self invocation (fires reach Claude as plain text)[^ScheduledTasks] — a real
   mechanism, not prompt courtesy [minority: lane-1/adversarial — the frontmatter mechanism;
   interactive-only is convergent]. Port plan: "human approves each step."[^PortPlan]
2. **Full-strength FEOV `/research`** on the stub's question, evidence base pinned, stub +
   provenance staged as inputs; the stub's alternatives become seed hypotheses for the
   frontier. Judgment tier intact (`judgmentModel` inherits the session model by design;
   "for keeper runs, omit `model` entirely"[^ResearchCommand]). This is where the $150–400
   class spend belongs: once per graduation, not once per day.
3. **On a VERIFIED (or human-accepted UNVERIFIED with rationale) report:** promotion
   proposal into `projects/` (or a `plans/` proposal in SDD shape with the plan-audit gate
   [minority: lane-3/local-probe — the plans/SDD path]) — a plan, not a rule change.
4. **Any change to rules/skills/commands/hooks lands as a pull request a human reviews and
   merges** — the loop never touches those paths even at graduation (§4's write fence has no
   graduation exception; the *human's interactive session* makes the edits, on the
   evidence). Even /graduate's own artifacts land in research/ideas/plans as PROPOSALS.

The DGM analogy is direct and is the design argument: DGM evaluates every change
empirically against a benchmark before archive entry, never on the proposer's say-so.[^DGM]
Honesty clause (round 2, R2-21, per the round-1 leaf read): DGM's ADMISSION bar is
compile-plus-retained-edit-ability — low scorers are deliberately retained in the archive
for open-ended exploration; benchmark performance orders selection, it does not gate
admission. Our /graduate is therefore STRICTER than DGM on exactly this dimension —
promotion is pass-required — so DGM is precedent for evaluation-before-entry and for
inspectable lineage, NOT for threshold-gated admission. Our
"benchmark" is the adversarial debate plus the human gate; /graduate is the validation
harness, and the git history of ideas/ → research/ → projects/ is DGM's
"transparent, traceable lineage" property — which is also what let DGM's authors *catch*
their agent faking test logs.[^DGMSakana] [minority: lane-2/primary-literature — the DGM
framing]

**Verdict on H2: SUPPORTED** — thin driver over existing machinery; picker mechanics cheap
and deterministic; judgment untouched. Confidence HIGH (internal cost evidence measured;
convergent primary literature; external stub-quality evidence MEDIUM).

---

## 3. H3 — Scheduling and headless reality (harness facts, checked at primary AND by live probe)

### 3.1 Verified live on this box [minority: lane-3/local-probe — both probes]

(2026-07-17, claude CLI 2.1.212.)

**Probe P1** — plain print mode with budget flag and JSON output:
`claude -p "Reply with exactly: OK" --model haiku --output-format json --max-budget-usd 0.10`
→ exit 0; JSON envelope carrying `result`, `is_error:false`, `num_turns`, `session_id`,
`total_cost_usd:0.0246903`, per-model usage breakdown, `permission_denials:[]`,
`terminal_reason:"completed"`.[^HeadlessProbe] Everything a scheduler needs for
machine-readable success/failure and spend accounting is in one JSON object on stdout.

**Probe P2** — plugin slash command, headless:
`claude -p "/prosthetic-conscience:probe" --model haiku --output-format json --max-budget-usd 0.30`
→ exit 0; the marketplace-installed plugin command loaded and ran; it spawned a subagent;
cross-plugin `skills:` preloading reached the subagent (the probe quoted the critical-stance
verification line verbatim); and — decisive for §4 — the probe's hook-test write was
**permission-denied** because no write permission was pre-configured, and the run reported
`Hook | BLOCKED — subagent write permission denied` in its result text rather than
attempting a workaround.[^HeadlessProbe] Cost: $0.058 at haiku. The H3 falsifier (headless
can't load plugin commands) is disproven by direct probe for the non-bare case.

### 3.2 What `claude -p` verifiably supports (primary docs, accessed 2026-07-17)

- Non-interactive full agent loop; **all CLI options work with `-p`**; "User-invoked skills
  and custom commands work in `-p` mode: include `/skill-name` in the prompt string and
  Claude Code expands it before running"; interactive-only built-ins (`/login`)
  excluded.[^HeadlessDocs]
- **Without `--bare`, `claude -p` loads the same context an interactive session would** —
  hooks, skills, plugins, MCP servers, CLAUDE.md, auto memory. `--bare` "skip[s]
  auto-discovery of hooks, skills, plugins, MCP servers, auto memory, and CLAUDE.md" and
  "will become the default for `-p` in a future release."[^HeadlessDocs] The scheduled
  recipe must be future-proofed now: either run non-bare explicitly from the repo root, or
  (the reproducible recipe, preferred by lanes 1+2) run `--bare` with everything explicit —
  `--plugin-dir` pointing at an **operator-owned, read-only plugin copy outside the write
  surface** (e.g. `~/.claude/sleeper/plugins/`, refreshed only by the human-gated plugin
  update path — NEVER the repo's `plugins/` working tree, which would make the repo copy
  the executing copy and void §4.3 layer 2's independence leg: round-1 fix, R1-15).
  **The copy's lifecycle is now BUILT, not assumed (round 2, R2-8 — a merged guardrail
  fix is not a deployed one; the round-1 pin minted a third plugin copy with no owner,
  no refresh path, and a preflight that asserted LOAD, not VERSION):** (a) creation and
  refresh are one named human action — `refresh-sleeper-copy`, a step in
  scheduling.md's guardrail-PR merge checklist: after any PR touching `plugins/`, the
  human copies the merged tree into `~/.claude/sleeper/plugins/` and records its content
  hash (directory SHA-256 over sorted file hashes) as the operator-approved value beside
  the ledger; (b) the wrapper preflight recomputes the copy's hash and FAILS CLOSED on
  mismatch with the approved value (tamper or partial refresh both abort — invariant 7);
  (c) staleness is a doctor line — "sleeper plugin copy: matches approved hash;
  approved hash recorded <date> vs last plugins/ merge <date>" — so a guardrail fix
  merged but never refreshed into the executing copy is VISIBLE, not silently believed
  deployed.
  `--mcp-config` + `--strict-mcp-config`, `--settings` for the permission profile. Bare mode also "skips OAuth and keychain reads":
  authentication must come from `ANTHROPIC_API_KEY` (or apiKeyHelper in `--settings`) — the
  recipe must document the API-key path and its billing consequences (§5) [minority:
  lane-2/primary-literature — the bare-auth consequence].
- **Load verification is machine-checkable**: `--output-format stream-json` begins with a
  `system/init` event carrying `plugins` (name+path) and `plugin_errors` arrays — the docs
  suggest failing CI when a plugin did not load. Step 0 asserts sleeper-service +
  frank-exchange-of-views are present and **aborts loudly if absent**.[^HeadlessDocs]
- **Background workflows are wait-capped at 10 minutes by default** (v2.1.182+,
  `CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS`; `0` = no limit). A debate workflow spawned inside
  a headless run can exceed 10 minutes routinely — the wrapper MUST raise this ceiling or
  the harness truncates paid-for work.[^HeadlessDocs] [minority: lane-1/adversarial]
- Exit behavior (requalified round 1, R1-6): code 0 on success — probe-corroborated
  (P1/P2); on error the CLI exits nonzero, but the cli-reference page publishes no
  exit-code table, so the exact value is unpinned — **the wrapper treats ANY nonzero exit
  as failure** rather than matching specific codes (`--max-turns` "exits with an error"
  per its flag text; the code value is unpublished).
  `--output-format json` carries `total_cost_usd` + per-model breakdown "so scripted
  callers can track spend per invocation"; `--json-schema` yields schema-conforming
  `structured_output` (invalid schema exits with error ≥2.1.205).[^HeadlessDocs][^CliReference]
- Permission pre-configuration headless: `--allowedTools`, `--disallowedTools`,
  `--permission-mode` (incl. `dontAsk` — auto-denies TOOLS not pre-approved via allow
  rules, **with one documented carve-out (corrected round 3, R3-14): a built-in,
  NON-configurable set of read-only Bash commands (`ls`, `cat`, `echo`, `pwd`, `head`,
  `tail`, `grep`, `find`, `wc`, `which`, `diff`, `stat`, `du`, `cd`, and read-only forms
  of `git`) runs without a prompt "in every mode"; the doc's stated remedy is an explicit
  `ask`/`deny` rule per command, which §4.2 now carries** — re-verified live at the
  permissions doc, blue round 3; documented
  for "locked-down CI runs").[^CliReference][^HeadlessDocs][^PermissionsDoc]
- **Session continuity for resume-after-abort:** `--resume <session_id>` from the same
  directory; the session id arrives in the JSON envelope.[^HeadlessDocs]
- **Workspace trust is interactive-only.** "In non-interactive mode with `-p`, no dialog
  appears and the rules stay ignored" — project-settings allow rules apply only after the
  workspace trust dialog was accepted interactively. Recipe precondition: trust the repo
  interactively once before scheduling (true on this box); a fresh clone scheduled cold
  silently runs with project allow rules ignored — fails SAFE (denied writes), but
  fails.[^PermissionsDoc]
- MCP headless: `--mcp-config <file-or-json>` + `--strict-mcp-config` (ignore all other MCP
  configs) — the determinism an unattended run wants; in non-bare mode the project
  `.mcp.json` loads as in interactive sessions.[^CliReference]
- 10MB stdin cap on piped input [minority: lane-2/primary-literature].[^HeadlessDocs]

### 3.3 Disconfirming pass — where headless breaks (all traps named)

1. **Headless regressions are a live, recurring class**: `claude -p` hung/appeared to hang
   on Windows during the slash-command scan for eight consecutive releases
   (v2.1.161–v2.1.168; fixed v2.1.169)[^WindowsHang] [minority: lane-1/adversarial]; custom
   slash commands were not discovered in CLI/SSH headless mode in earlier versions (#14246,
   #837).[^SlashHeadlessIssues] The current doc's statement supersedes them, but a live
   acceptance test must remain — doc-says is not run-verified (probe P2 covers the
   non-bare case on 2.1.212; the bare+--plugin-dir case is still open). **The Phase-4
   acceptance test is RESTATED two-legged round 5 (R5-2 — the port plan's verify step,
   "Headless `claude -p \"/self-improve\"` produces a run dir + idea stub; touches only
   research/+ideas/"[^PortPlan], is unsatisfiable as written under the R4-1 trampoline:
   the command markdown carries `disable-model-invocation: true` and NO payload, so the
   fire reaches Claude as inert instruction text and produces no run dir — and the
   cheapest way to make the OLD printed gate pass would be to re-inline the payload,
   undoing R4-1; the plan's step is superseded as-written, not inherited):** leg (i) —
   `node sleeper-wrapper.mjs --manual` produces the run dir + idea stub, touching only
   research/ + ideas/ (the capability leg, now aimed at the artifact that actually hosts
   the loop); leg (ii) — `claude -p "/self-improve"` produces NO run dir and NO write
   (trampoline INERTNESS, the verifiable R4-1 property — a passing leg (ii) is evidence
   the payload has not crept back into `commands/`). **scheduling.md ships a preflight** (trivial
   `claude -p --output-format json` ping asserting plugin presence, plus a version pin
   known-good for headless) and treats silent success as unproven — the routines doc itself
   warns green run status ≠ task success [minority: lane-1/adversarial — the
   green-status warning].[^RoutinesDocs]
2. **MCP under headless is the fragile joint** [minority: lane-3/local-probe — the issue
   trio]: three relevant upstream defects — TWO leaf-checked OPEN 2026-07-17: stdio MCP
   tools silently missing on the first turn when server startup exceeds ~2s (regression
   since 2.1.144, #76239); stdio tool calls hanging with several servers loaded, worked
   around by `--strict-mcp-config` (#68375) — and ONE historical: HTTP MCP `-p` runs
   exiting silently (#32191, 2.1.58–2.1.71 era, **CLOSED as duplicate**, canonical
   untraced; corrected round 1, R1-5 — the round-0 "leaf-checked OPEN" claim was false for
   this issue; the phenomenon class stays on the checklist, the OPEN status does
   not).[^McpHeadlessBugs] Open ≠ will-be-fixed: the design owns the workaround —
   (a) the loop's MCP profile is `--strict-mcp-config --mcp-config <sleeper-mcp.json>`
   naming **qmd only** (fewer servers is the #68375 mitigation). **Round-2 correction
   (R2-17): the round-1 parenthetical "research subagents reach pdf/arxiv tools via
   ToolSearch" was false under this flag — ToolSearch discovers deferred tools only from
   DECLARED servers, and `--strict-mcp-config` ignores the project `.mcp.json` that
   declares pdf-reader/arxiv-latex, so nothing declares them nightly. The trade is
   chosen, not straddled: the nightly profile ACCEPTS degraded PDF citation capability
   (rather than naming three servers and re-opening the #68375 multi-server hang class
   in exactly the unattended mode with nobody watching), which makes the
   research-protocol PDF MUST-try clause structurally unsatisfiable in nightly runs —
   stated as a standing confidence caveat in every stub (§2.3), with full PDF tooling
   restored at human-present graduation where the claims are re-derived anyway.
   Consistent with the tiering doctrine: what is cheapened nightly is citation-polish
   mechanics on a proposal artifact, never the judgment that promotes it; (b) qmd is addressed via
   the **HTTP daemon**, verified live on this box 2026-07-14 (`qmd mcp --http --daemon`,
   PID file, `/health`, MCP Streamable HTTP at `:8181/mcp`)[^QmdDaemon] — avoiding the
   per-invocation model-load penalty (bare-CLI hybrid query 36.3s vs daemon lex 2.9s); the
   scheduler preflight curls `/health` and starts the daemon if absent, converting the
   slow-stdio-start class into a non-event; (c) the loop **degrades to Grep/Read** over the
   corpus when qmd is unreachable — the efficiency run's seats did exactly this and
   produced usable work, logging the absence as friction.[^QmdFallback]
3. **/loop is NOT the scheduling substrate**: session-scoped, fires only while the session
   is open and idle, recurring tasks expire after seven days, "no catch-up for missed
   fires."[^ScheduledTasks] Fine for babysitting a PR; wrong for a daily standing loop.

### 3.4 The scheduling ladder (docs/scheduling.md content)

| Rung | Mechanism | Fit |
|---|---|---|
| 0 | Manual — the human invokes the SAME wrapper by hand (`node sleeper-wrapper.mjs --manual`); the loop PAYLOAD lives in the wrapper's phase-1 prompt (sourced from skills/continuous-learning), and `/self-improve` is a THIN TRAMPOLINE that only tells a human to run the wrapper — `disable-model-invocation: true`, no payload, inert to scheduled/model fires (rewritten round 5, R5-2: the prior cell still described the round-3 payload-in-command shape R4-1 closed — a builder reading it would re-create the payload-carrying invocable command, the exact vector) | always available; the default until the human opts into scheduling (port plan resolved decision 6: daily cadence, scheduling always human-opt-in)[^PortPlan] |
| 1 | **OS scheduler + `claude -p`** (Windows Task Scheduler on this box; cron/systemd elsewhere) | **RECOMMENDED default among SCHEDULED rungs, once the human opts in** (rung 0 remains the overall default until then — §1.4, R2-15) — the only local option where every flag is explicit and version-pinnable; full local files, qmd daemon, plugins, hooks; machine must be on; the launch script (Node, mechanics) owns preflight + budget ledger + idempotence + JSON-result logging |
| 2 | Desktop scheduled task | persistent, visual schedule, local files + MCP config, per-task permission config; requires the Desktop app machinery[^ScheduledTasks] |
| 3 | **Cloud Routine** (research preview) | off-box alternate: runs on Anthropic infra with the machine off; schedule ≥1h interval; **fresh clone of the default branch, no local files; no permission prompts (fully autonomous); pushes restricted to `claude/`-prefixed branches by default** — a structural human-gate (a cloud sleeper can only ever produce a branch a human merges); MCP via connectors or committed `.mcp.json`; counts against subscription usage + a daily routine-run cap; requires claude.ai login (API-key auth unsupported [minority: lane-1/adversarial — the auth detail])[^RoutinesDocs] |
| 4 | GitHub Actions `schedule` trigger | CI-grade recipe; secrets-managed (`ANTHROPIC_API_KEY`); the port to team contexts. Best-effort semantics: runs commonly delayed 5–30 min at high-load windows and can be dropped under queue saturation; 5-min minimum interval; no SLA [minority: lane-2/primary-literature — the delay/drop detail].[^GhaSchedule] Acceptable for a daily job — a dropped day is harmless by design (idempotence); forgoes the local qmd daemon |

**Missed-run tolerance is the load-bearing scheduler option on a machine that sleeps**
[minority: lane-2/primary-literature]: Windows Task Scheduler "Run task as soon as possible
after a scheduled start is missed"; systemd timer `Persistent=true`; anacron
interval-since-last-run semantics. Cron alone is the wrong default on any machine that
sleeps; the recipe should say so.[^MissedRun]

**Idempotence rule** (makes every scheduler failure mode harmless): the run is keyed by date
(`research/<date>_self-improve/`); on start, if today's run dir exists and is complete, exit
0; if incomplete, resume it (`--resume` from the recorded session id) instead of starting
fresh. A dropped day is a no-op; a double fire is a resume. **Resume cap (round 1,
R1-29): k=3** — after three failed resume attempts on one run dir the wrapper writes a
`DEAD` marker (in the OPERATOR-OWNED dir beside the ledger, dated — never in the
loop-writable run dir; and the exit-0/resume/fresh decision reads the wrapper's OWN
step-7 ledger record for completeness, never run-dir contents: R2-12, closing the sibling
R1-19's own text had named) plus a dated abort record, and the next fire mints a fresh
dir, so a wedged run (corrupt dir, dead session id, deterministic abort) cannot be
resumed nightly forever **within one run dir**. Scope narrowed honestly round 2 (R2-10):
the k=3 cap bounds per-DIR, not per-CAUSE. **In-night retry semantics, stated, and the
arithmetic recomputed (round 3, R3-12 — both round-2 figures were wrong as printed):**
the wrapper fires ONCE nightly and performs ONE resume attempt per fire — k=3 accrues
ACROSS nights, never within one invocation — so the per-night worst case is ≤1×$5 (one
ceiling-priced session), not k×$5; a deterministic root cause burns ≤$5/night. **The
cap/HALT interaction recomputed round 4 (R4-10 — the round-3 "night 12 OR the cap,
whichever comes first" race was wrong at ceiling pricing):** at the $5 ceiling, deaths land
nights 4 and 8 (initial + 3 resumes = 4 nights per dir); the ledger preflight skips from
~night 11 (≥$50 after 10×$5); nights 11–30 are CAP-SKIPS during which NO death accrues
(the run does not launch), so the third same-signature death CANNOT occur in month 1 and
the M=3 HALT lands early the NEXT month (~$5–10 into month 2, after the cap un-pre-empts at
rollover). There is no "whichever comes first" race — the cap ALWAYS pre-empts first
(~night 10), pauses death accrual with it, then releases at month rollover, and the HALT
follows. Worst-case deterministic-cause burn ≈ one monthly cap + ≤2 nights of the next
month (~$55–60 across two months). The bounded-by-the-cap conclusion survives; only the
printed race and single-month bound were wrong. **Per-cause HALT (R2-10):** the wrapper
normalizes each death's abort reason into a dead SIGNATURE — **normalization SPECIFIED
(round 3, R3-11; the fresh-dir mechanism guarantees per-night-unique paths and dates, so
an identity-keyed signature over the raw first line would never match twice and the HALT
would never fire while behaving exactly as specified — the corpus's
identity-keyed-detector-lineage-blind lesson applied to the design's own repair):**
`signature = exit class + abort-reason TEMPLATE`, where the template is the first
abort-record line with dates, paths, session ids, nonces, and all digit runs replaced by
placeholders (`<date>`, `<path>`, `<id>`, `<n>`), plus the error code where one exists —
recorded beside the DEAD marker; after M=3 consecutive fresh-dir deaths with
the same signature the wrapper writes a `HALT` marker — preflight refuses to launch
until the human clears it (§2.2 step 0) — and raises the dead-man flag with the
signature as the printed reason. The doctor line additionally prints
HALT-firings-since-last-clear INCLUDING zero (R3-11): zero firings across accumulating
DEAD marks is evidence the normalization is broken, not evidence of health. One residual
owned in a clause: deaths ALTERNATING between two causes never show M=3 consecutive
same-signatures — the per-signature counts on the doctor line (R3-6) are the
human-visible backstop for that shape. **Dead-man surface (R1-29; re-argued round 2, R2-9):**
"loudly" gets a defined reader — the wrapper maintains `last-successful-run` plus the
last skip/abort REASON (R2-18) in the
operator-owned dir, and `/prosthetic-conscience:doctor` reports "last successful sleeper
run: N days ago (last skip reason: <reason>)" PLUS per-signature skip/abort counts since
the last human clear, the HALT-firings count, **and a qmd-degrade streak count (R4-7 —
"qmd degraded on M of the last N runs"; a chronic daemon-start failure that silently
degrades recall otherwise surfaces nowhere, since a qmd degrade is neither a skip nor an
abort reason)** (R3-6/R3-11 — TAMPER and HALT persist
until cleared; every other class would otherwise surface only as most-recent and scroll
off, leaving the §1.5 stub flag the sole standing channel). Round-2 honesty (R2-9): doctor is
PULL-only, and the loop's adoption motive removes exactly the babysitting that made "the
surface the human already looks at" an assertion rather than evidence. Decision, both
horns owned: (a) a cheap push-adjacent surface is added — the same staleness check runs
as a SessionStart-hook warning line in every INTERACTIVE session on this box once
`last-successful-run` exceeds N=7 days (one file-stat; a channel the operator passively
receives whenever they use Claude Code at all — invariant 7's human surface — and one
that is independent of the sleeper scheduler, so scheduler death does not silence it).
The warning is CONDITIONED on scheduling-enabled — a flag the operator's schedule-setup
step sets and the disable step resets (round 3, R3-10: an operator who schedules then
deliberately stops must not be nagged forever, or the warning stops being read) — and it
always prints its last-check timestamp as a liveness signal (R3-9: a silent hook and a
passing hook must not look identical). **Two clauses close the disarm-custody gap round 4
(R4-14):** (i) the scheduling-enabled flag lives in the OPERATOR-OWNED dir beside the ledger
(`~/.claude/sleeper/`), OUTSIDE every session write surface — so the loop cannot silence its
own dead-man channel by flipping the flag (the guards-modifying-guards class §4 forbids);
(ii) the SessionStart line is NEVER fully silent — when scheduling is disabled it still
prints ONE token, `"sleeper: scheduling disabled; last check <ts>"`, so a deliberately
disabled loop and a SILENTLY-DISARMED one (flag LOSS via misconfig or uninstall/reinstall —
the modal marker-loss drift R1-28 treats as certain) are DISTINGUISHABLE: flag-absent prints
`"sleeper: flag missing; last check <ts>"` (an anomaly), flag-off prints the disabled token
(a choice), and flag-on-stale prints the staleness warning. **Flag-absent is gated round 5
(R5-7 — flag-absent is also the DEFAULT state of every install that never ran
schedule-setup, and printing it as a per-session anomaly on such installs is the
alert-fatigue mode this report's own Dependabot evidence catalogs, aimed at the design's
own dead-man channel): the anomaly print fires ONLY when a prior `last-successful-run`
record exists in the operator-owned dir — flag missing WITH run history = something was
lost (anomaly, print every session until resolved); flag missing with NO run history =
never configured (silent, or at most a one-time notice) — so never-configured and
flag-lost are distinguished by the same durable record the dead-man surface already
maintains.** Silent-disarm ≠ disabled-by-choice, re-closing what R3-9's always-print
clause established for the scheduling-enabled case;
(b) the residual stated plainly — if the operator uses neither doctor nor interactive
sessions on this box, death-detection latency is unbounded; accepted for a
single-operator box whose operator's daily working tool this is, and §6 row 15's grade
now carries that latency term. "Missed day (harmless)" and "loop dead three weeks"
remain distinguishable without a new monitoring channel.
Persistent death is §6 row 15.

**Rung-3 caveats stated honestly:** the qmd daemon does not exist in the cloud environment
(setup-script install possible — ~973MB models — but unverified; degrade to lex-only or
repo-committed `.mcp.json` stdio); routine behavior/limits are research-preview volatile
("Behavior, limits, and the API surface may change"); all claude.ai connectors are included
by default and must be trimmed per routine [minority: lane-3/local-probe — the connector
default]; and everything a routine does appears as the human's GitHub identity — an argument
FOR the PR gate, not against the rung.[^RoutinesDocs]

**Per-rung gate survival (added round 1, R1-27 — the §4 gate stack is a rung-1 artifact,
and docs/scheduling.md carries this table).** §4.2's permission profile and §5.1's wrapper
machinery bind rungs 0–1 fully (rung 0 since the R3-2 `--manual` decision — the round-2
table's rung-0 SPLIT cell is superseded) and rung 2 at most partially; at rungs 3–4 most
layers do not exist unless rebuilt for that environment. **Climbing past rung 2 is itself a
graduation-grade decision** (its own idea stub, its own human gate) — never a config
toggle: as documented round 0, an operator climbing the ladder would reach the
maximum-blast-radius configuration (fully autonomous, connectors default-included, none of
§4.2) still believing the gates were on.

| Gate / control | R0 manual | R1 OS + `-p` | R2 Desktop task | R3 Cloud Routine | R4 GitHub Actions |
|---|---|---|---|---|---|
| L1 permission profile (operator-owned `--settings`) | **YES** — the wrapper passes the same `--settings` profile (R3-2) | **YES** | PARTIAL (per-task permission config; `--settings` passthrough unverified — OQ11/OQ19) | **NO** — no permission prompts; the profile file is absent from the fresh clone, and committing it moves it inside the tree the loop's PRs touch | PARTIAL — a workflow-committed settings file sits inside the PR-touchable tree (rung-4 inversion, R1-27) |
| L2 hook fence + step-0 denial canary | **YES** — the R2-16 SPLIT is dissolved (R3-2): rung 0 passes through the wrapper, so canary + fired-record run manually too | **YES** (wrapper canary + fired-record — R2-2) | PARTIAL (canary needs the wrapper) | only if hooks load from committed plugin config; canary absent | same as R3 |
| Wrapper controls (preflight, ledger, idempotence, resume cap, snapshots) | **YES** — identical wrapper, `--manual` (R3-2) | **YES** | NO (Desktop owns scheduling; wrapper absent unless the task invokes it) | **NO** | PARTIAL (rebuildable as workflow steps) |
| Provenance/origin-tag + corroboration gate (§1.5 — origin markers, window log, memory-surface snapshot diff, harvest-side corroboration gate) — added round 4 (R4-13: the poisoning gate is a wrapper-hosted control, so it is void at exactly the rungs the wrapper is PARTIAL/NO; a markerless sleeper dir reaching the local corpus is counted NON-sleeper next harvest — SUPERSEDED round 5 by the R5-3 positive pool: a markerless post-epoch dir is QUARANTINED, never counted) | **YES** — wrapper stamps at creation | **YES** | PARTIAL (wrapper absent → sleeper dirs land markerless → the R5-3 positive pool QUARANTINES them — they can no longer corroborate — but sleeper-origin ATTRIBUTION and the 1-per-class cap still need the rebuilt stamp, so adoption still requires it) | **NO** (fresh clone, no wrapper, no local harvest; R5-3 quarantine protects the local corpus from anything synced back markerless) | NO unless rebuilt as workflow steps + a harvest that reads the CI origin marker (R5-3 quarantine bounds the damage of NOT rebuilding: markerless CI dirs count for nothing rather than as corroboration) |
| L5 git/PR gate (server-side, branch protection + CODEOWNERS) | YES | YES | YES | YES — plus native `claude/`-branch push restriction | YES |
| L6 `disable-model-invocation` on /graduate | YES | YES | YES | YES (doc-verified for scheduled fires) | YES |
| Per-run budget (`--max-budget-usd`) | YES | **YES** | UNKNOWN (per-task config — OQ19) | platform caps only (subscription draw-down + daily routine cap) | YES (flag in workflow) |

Reading: rungs 0 and 1 are the only rungs where every layer is live as designed (rung 0
joined round 3 via the `--manual` wrapper — R3-2), which is *why* rung 1 is
the recommended default among scheduled rungs, not merely a convenience preference.
Rungs 3–4 keep the
load-bearing layer (5) and shed most of the rest; adopting them means re-deriving the
stack for that environment, documented in the adoption stub — **which MUST include the
provenance/origin-tag + corroboration gate (R4-13): a rung that produces sleeper artifacts
without the wrapper's origin stamping poisons the next harvest, so the adoption stub's
graduation-grade requirement names rebuilding this gate (a CI-side origin marker its harvest
reads) as a precondition, not an optional layer.** **Rung-0 execution shape,
DECIDED (round 3, R3-2 — superseding the R2-16b out-of-ledger accept, and the round's
priority fix):** rung 0 is NOT a wrapper-less mode. Round 2 left a contradiction as
printed — §1.4 said "same code path" while this section said manual runs "do not pass
through the wrapper" — which meant the DEFAULT and possibly terminal mode had no executor
for four of eight steps (0/2/4/7 are wrapper-hosted after R1-16/R2-3/R2-4), and its run
dirs landed markerless: untagged (evading the §1.5 1-per-class cap) AND non-sleeper
(able to satisfy the corroboration gate) — the R2-5 circumvention shape recurring in the
design's default mode. R2-16b's own closing sentence named the cheap fix; round 3 adopts
it as the only shape under which "same code path" is true: **a manual run invokes the
SAME wrapper with `--manual`** — identical phase sequence (preflight, canary, staging,
snapshots, ledger, record), the only differences being console-visible output and an
`origin: manual` ledger annotation. Consequences: steps 0/2/4/7 have their executor in
the default mode; manual spend is IN-ledger, so the R2-16b out-of-ledger accept and its
mixed-month undercount are DISSOLVED, not accepted; and the R2-5 origin marker + R2-6
corroboration gate hold at rung 0 — a manual run's run dirs are stamped at creation like
any scheduled run's (§6 row 10 re-affirmed). An operator bypassing the wrapper by pasting
the command markdown into a bare interactive session is out-of-contract. Round 4 (R4-1)
makes this harder to reach AND corrects a polarity error: (a) the payload no longer lives
in `commands/` — `self-improve.md` is a thin trampoline carrying `disable-model-invocation:
true` (matching `graduate.md`), so a scheduled/model fire reaches Claude as inert text and
there is no invocable command carrying the loop payload to paste; the paste path requires a
human deliberately copying the wrapper's phase-1 prompt out of the skill file. (b) **Polarity
correction:** the round-3 text said a markerless out-of-contract dir's friction "is
ordinary-classed, requiring corroboration" — that is BACKWARDS. Harvest tags only MARKED
dirs `origin: sleeper`; a markerless dir is classed NON-sleeper, so its friction does not
NEED corroboration — under the round-4 default-negative pool it PROVIDED the non-sleeper
corroboration that ordinary sleeper-only classes wait on (the R2-5 origin-tag-escape
vector, 4th surface). **SUPERSEDED round 5 (R5-3): under the positive-pool root invariant
(§1.5) a markerless post-epoch dir provides NOTHING — it is quarantined
(counts-for-nothing) and surfaced `unattributed` for human attribution, so the paste-run
corroboration residual this paragraph priced is DISSOLVED, not risk-accepted.** The other
containment legs stand: no marker ⇒ the fence never engages AND the interactive profile's
own gates apply; and its dirs are swept to `origin: sleeper` by the window log if any
wrapper run brackets them (bounded by R4-6's next-START window close), so a paste-run
overlapping a scheduled window is caught. OQ24 — the round-4 deferral of a loop-SHAPED-dir
quarantine heuristic — is RESOLVED by promotion in generalized form: the built quarantine
keys on missing positive provenance, not on schema-matching a dir's shape, so it pays no
false-positive classification on ordinary research dirs (an unattributed human dir is
surfaced for one keystroke, never misclassified in either direction).

**Claude Code on the web** (cloud sessions) [minority: lane-2/primary-literature]: isolated
Anthropic-managed VM per session, credentials held outside the sandbox — attractive
isolation, but background-task agents are documented failing there (sandbox recycled between
turns, #23707 — CLOSED NOT PLANNED, leaf-confirmed by red round 1: the platform will not
fix it, which strengthens the not-the-default call) and there is no scheduling primitive.
Not the Phase-4 default; revisit when a scheduled-trigger primitive exists.[^WebSandbox]

**Verdict on H3: CONFIRMED WITH CONDITIONS** — every needed capability is documented
first-party for `-p` and the core path is live-probe-verified; the design must carry (a) the
load-verification preflight, (b) the background-wait ceiling, (c) a version pin +
known-regression note, (d) the `--bare` default flip on the horizon, (e) the MCP-headless
workaround set. Confidence SPLIT (round 1, R1-20): **HIGH for the probed non-bare path**
(probe P2, CLI 2.1.212); **OPEN for the preferred `--bare --plugin-dir` recipe** — the
one-command bare probe is the Phase-4 acceptance test (open question 1), and until it
runs, the recipe's recommendation stands on doc evidence only. A verdict stamp must not
grade higher than its recommended configuration's evidence.

---

## 4. H4 — Consent gates that autonomy cannot erode (SUPPORTED; the permission-rules-only variant REFUTED at the leaf node)

### 4.1 Disconfirming pass first: why instructions are not gates

- **Local evidence — instructional gates degrade under load:** a prompt-level MUST (PDF
  extraction try-before-grade-down) was skipped live by a lane ("the MUST-try clause has no
  observable"); batching prose was ignored wholesale at the bulk tier (0/175 multi-call
  messages).[^FrictionRun4][^Backlog] The corpus's own conclusion is the script-vs-prose
  doctrine.[^ResearchCommand] "Please only write research/ and ideas/" is a wish, not a
  gate.
- **Platform doctrine, verbatim:** "Permission rules are enforced by Claude Code, not by
  the model. Instructions in your prompt or CLAUDE.md shape what Claude tries to do, but
  they don't change what Claude Code allows."[^PermissionsDoc]
- **Permission rules are empirically leaky for exactly the file-protection case** [minority:
  lane-1/adversarial — the issue trio]: Edit/Write tools bypassing `permissions.ask` rules —
  reported specifically as "impossible to protect critical files (hooks, settings)" — is
  **closed NOT PLANNED** upstream (#22055)[^PermAskBypass]; deny rules non-enforced for
  Read/Write tools was reported and persisted after a claimed fix (#6631, closed; reporter
  re-confirmed at v1.0.93)[^DenyRWIssue]; Bash deny-rule non-enforcement was reported and
  closed as duplicate (#25621).[^DenyBashIssue] Per red's Pattern A discipline these
  closures invert the fix story: **the design must own the gate; upstream will not.**
  (Statuses verified by direct fetch 2026-07-17; behavior in current 2026 builds unverified
  — but a gate you must re-verify per release is not a foundation.)
- **Autonomous loops editing their own constraints is a realized failure, not a
  hypothetical:** Sakana's AI Scientist, given write access to its own experiment harness,
  edited its own execution script — relaunching itself in an infinite loop and extending its
  own timeout; the authors' stated lesson is sandboxing/isolation, i.e., mechanism outside
  the model.[^AIScientist] DGM's agent, asked to fix its own log-faking, "removed the
  markers we use in the reward function to detect hallucination (despite our explicit
  instruction not to do so)" — caught only because the archive lineage was externally
  inspectable [minority: lane-2/primary-literature].[^DGMSakana] STOP measured it: across
  10,000 sampled improver generations, GPT-4 disabled the provided sandbox flag in 0.42%
  (CI 0.31–0.57%) of cases, and **adding an explicit safety warning did not statistically
  significantly reduce it** (0.46% with warning — in fact insignificantly HIGHER; figures
  re-pinned round 1 at ar5iv §6.2/Table 2 by three independent red lenses, OQ8 resolved)
  [minority: lane-2/primary-literature].[^STOP]
- **Security practice converges:** OWASP's Excessive Agency entry prescribes least
  privilege, human-in-the-loop approval for high-impact actions, and authorization enforced
  in systems external to the LLM.[^OWASP] The AI-control line (Greenblatt et al., ICML
  2024) evaluates safety protocols under the assumption the model **intentionally subverts**
  them — the correct design stance for gates that must survive an optimizing loop, even one
  we trust today [minority: lane-2/primary-literature].[^AIControl]
- **The deny path works headless and is honestly reported** — probe P2's write denial
  (§3.1) [minority: lane-3/local-probe].

### 4.2 Write surface and the permission profile

**Write surface: `research/**` and `ideas/**` only. Deny-set: `plugins/**`, `.claude/**`
(settings, hooks wiring), `.claude-plugin/**`, `.mcp.json`, `CLAUDE.md`, `MEMORY.md`,
`.github/**`, `plans/**`, `projects/**`, `**/hooks.json`, `**/settings*.json`, the sleeper
guard's own binary/manifest paths, and `~/.claude/**` (absolute-anchored).**

The scheduled invocation runs `--permission-mode dontAsk` with an **operator-owned settings
file outside the repo working directory** (lane-2: e.g. `~/.claude/sleeper/`, passed via
`--settings`; lane-3: deny set at user scope in `~/.claude/settings.json` — both placements
put the defining file outside the loop's writable universe). Round-1 corrections baked
into the sample: `disableBypassPermissionsMode` moved INSIDE the `permissions` object —
the doc scopes the key there, and at top level it is silently ignored, so the advertised
lockout never engages (R1-8, config-sample-schema class); all guardrail rules written
`//`-absolute — bare-relative rules anchor at the process CURRENT DIRECTORY, and a
scheduled wrapper that forgets to `cd` (Task Scheduler's default start-in is System32)
would silently relocate a relative deny set in exactly the unattended mode it exists for
(R1-12); `Bash(node scripts/*)` removed — it was a write-capable subprocess channel inside
the allowlist AND its pattern matched repo-root `scripts/`, not the actual
`plugins/sleeper-service/scripts/` layout (R1-16; harvest now runs wrapper-side, §1.3);
Read/Grep/Glob allow-scoped to the repo plus the read-only plugin copy — unscoped Read
over the box (including `~/.claude` session transcripts and credentials) combined with
model-authored egress was an unexamined exfiltration channel (R1-17); WebSearch removed
from the nightly default (injection-through-retrieval, R1-18 — human-present graduation
runs keep it):

```jsonc
// sleeper-permissions.json — lives OUTSIDE the repo, operator-owned.
// <REPO> = //c/Users/gbloc/Projects/special-circumstances — write the literal absolute
// form in the real file; shortened here for column width (R1-12: never bare-relative).
{
  "permissions": {
    "allow": [
      "Read(<REPO>/**)", "Grep(<REPO>/**)", "Glob(<REPO>/**)",
      "Read(//c/Users/gbloc/.claude/sleeper/plugins/**)",  // the read-only executing plugin copy (R1-15)
      "Edit(<REPO>/research/**)", "Edit(<REPO>/ideas/**)",
      "Bash(git status)", "Bash(git diff)", "Bash(git log)",
      "Bash(git log --oneline -20)",                            // EXACT argv, wildcard dropped (R3-15: "git log *" licensed
                                                                //   --output=<path> — a git-NATIVE arbitrary-file write,
                                                                //   reproduced live: exit 0, file created). R5-1: these four
                                                                //   rules are DEAD under the bare `Bash` deny below — deny
                                                                //   supremacy + bare-name removal take the tool out of
                                                                //   context entirely, so no Bash allow rule NOR the built-in
                                                                //   read-only-git carve-out can fire (the round-3 comment
                                                                //   "the carve-out auto-approves read-only git regardless"
                                                                //   was REFUTED by the same doc's bare-deny semantics and is
                                                                //   withdrawn). Retained ONLY as declared rung-rebuild
                                                                //   intent for environments (rungs 3-4) built without the
                                                                //   bare deny — where the carve-out DOES bind and these
                                                                //   become live declared-intent-not-enforcement again
      "Workflow",                                               // debate.js via the harness Workflow tool (R2-3); scriptPath
                                                                //   pinned to the read-only plugin copy by the sleeper-guard
                                                                //   hook (native rule-level scoping is OQ20)
      "mcp__qmd__*", "WebFetch(domain:arxiv.org)"               // extend deliberately, per need; NO WebSearch nightly (R1-18)
    ],
    "deny": [
      "Bash",                                           // R4-3 STRUCTURAL CLOSE (round 4): the
                                                        //   STRONGEST shape per invariant 6 — a BARE
                                                        //   tool name removes the tool entirely
                                                        //   (doc-verified). §2.2's session steps never
                                                        //   invoke Bash (git/node run WRAPPER-side), so
                                                        //   the whole dontAsk read-only-Bash carve-out
                                                        //   CLASS — enumerated members AND unlisted
                                                        //   members (sort/file/readlink/strings/less —
                                                        //   the doc set is "include[s]", non-exhaustive)
                                                        //   AND read-only git AND every git-native write
                                                        //   gadget — is closed at once, not neighbor by
                                                        //   neighbor. Everything below is now BELT: it
                                                        //   binds only where a rung rebuild drops the
                                                        //   bare deny (rungs 3-4) or the profile drifts.
      "Edit(<REPO>/.claude/**)", "Edit(<REPO>/plugins/**)", "Edit(<REPO>/CLAUDE.md)",
      "Edit(<REPO>/MEMORY.md)", "Edit(<REPO>/.mcp.json)", "Edit(<REPO>/.github/**)",
      "Edit(<REPO>/plans/**)", "Edit(<REPO>/projects/**)",
      "Edit(**/hooks.json)", "Edit(**/settings*.json)",
      "Read(//c/Users/gbloc/.claude/settings*.json)",   // belt: Read deny also blocks Edit (v2.1.208) — R1-17;
                                                        //   AND extends to recognized Bash file commands
                                                        //   (cat/head/tail/sed) per the doc's deny-reach
                                                        //   clause (R4-5) — so these named targets were
                                                        //   never Bash-cat exfiltrable even pre-bare-deny
      "Read(//c/Users/gbloc/.claude/.credentials*)",
      "Read(//c/Users/gbloc/.claude/projects/**)",      // session transcripts: the high-value exfil target
      "Read(//c/Users/gbloc/.ssh/**)", "Read(//c/Users/gbloc/**/.env)",  // R4-3: un-named box-secret
                                                        //   classes belt-denied too (the Read-deny/
                                                        //   Bash-file-command reach binds these even if
                                                        //   an unlisted read-only command survives)
      "Edit(//c/Users/gbloc/.claude/**)",
      // BELT (round 4: superseded as the PRIMARY close by the bare `Bash` deny above; RETAINED
      // as defense-in-depth for rung-migration environments that rebuild without the bare deny,
      // and against profile drift). R3-14: dontAsk's documented carve-out auto-approves a
      // built-in, NON-configurable read-only Bash set "in every mode"; the doc's own remedy is
      // per-command deny rules (deny at ANY level can never be allowed at another). The doc set
      // is stated with "include[s]" — NON-EXHAUSTIVE (R4-3) — so enumeration alone cannot be the
      // load-bearing close; the bare deny is. Enumerated members:
      "Bash(cat *)", "Bash(head *)", "Bash(tail *)", "Bash(grep *)", "Bash(find *)",
      "Bash(ls *)", "Bash(ls)", "Bash(diff *)", "Bash(stat *)", "Bash(du *)",
      "Bash(wc *)", "Bash(which *)", "Bash(echo *)", "Bash(pwd)", "Bash(cd *)",
      "Bash(sort *)", "Bash(sed *)", "Bash(file *)", "Bash(readlink *)",  // R4-3: doc-named
      "Bash(strings *)", "Bash(less *)",                                   //   classifier-reasoned + likely siblings
      // R3-15/R4-2 belt: git-native write flags in any command string — now incl. the `-o`/`-O`
      // SHORT forms (R4-2: `-o` matched none of the three long-form denies; `git format-patch
      // -1 -o /tmp/probe` → exit 0, arbitrary out-of-repo patch, leaf-verified) and the
      // subcommand writer family (R4-2/R4-15). The sleeper-guard hook's git-read ALLOWLIST is
      // the enforcement of record — §4.3 layer 4:
      "Bash(* --output=*)", "Bash(* --output *)", "Bash(* --output-directory*)",
      "Bash(* -o *)", "Bash(* -O *)",
      // R5-5: this belt is KNOWN-INCOMPLETE on ATTACHED short forms — git (like most
      // getopt short options) accepts `-o<value>` with no space, and `git format-patch -1
      // -o/tmp/x HEAD` exits 0 writing an out-of-repo patch while matching NO pattern
      // above (leaf-reproduced twice, red round 5). Deliberately NOT extended again: the
      // enumerate-and-extend approach cannot close the class (this escape recurred one
      // lexical form deeper INSIDE the round-4 extension itself). The closes of record
      // are the bare `Bash` deny (session + seats) and the hook's git read-ALLOWLIST
      // (rebuilt rungs); OQ18(c) leaf-tests the hook's attached-form handling.
      "Bash(git format-patch *)", "Bash(git archive *)", "Bash(git bundle *)",
      "Bash(git config *)", "Bash(git gc *)", "Bash(git repack *)", "Bash(git maintenance *)",
      "Bash(git push *)", "Bash(git commit *)"
    ],
    "disableBypassPermissionsMode": "disable"           // INSIDE permissions — R1-8
  }
}
```

(Primary read protection is the allow scoping itself — under `dontAsk`, the Read TOOL
outside the allow set is auto-denied (the R3-14 carve-out is a BASH carve-out; it does
not extend tool-level reads), and the `Read(...)` deny lines are belt on the named
high-value targets; the BASH read channel — which the carve-out left open — is closed
STRUCTURALLY round 4 by the bare `Bash` deny (R4-3), with the enumerated carve-out denies
above retained as belt for rung-migration environments and profile drift. Even before the
bare deny, the named `Read(...)` denies extended to recognized Bash file commands
(`cat`/`head`/`tail`/`sed`) per the doc's deny-reach clause (R4-5), so those targets were
never Bash-`cat` exfiltrable; the bare deny closes the residual UN-named-target hole R4-3
named. `disableAutoMode`, the same escape-hatch class as
`disableBypassPermissionsMode`, is a candidate addition — its key name/scope must be
leaf-verified against the current permissions doc at build before this report asserts it:
open question 17.)

Why this shape is sound rather than merely asserted:

- `dontAsk` auto-denies TOOLS not pre-approved — **corrected round 3 (R3-14): the round-2
  "auto-denies anything not allow-listed" was a closed-world premise the permissions doc
  refutes.** A built-in, NON-configurable read-only Bash set (`ls`, `cat`, `echo`, `pwd`,
  `head`, `tail`, `grep`, `find`, `wc`, `which`, `diff`, `stat`, `du`, `cd`, read-only
  git forms) "runs without a permission prompt in every mode" — so the default was
  closed for tools but NOT for that Bash subset. **Example re-pointed round 4 (R4-5):** the
  round-3 example `Bash(cat //c/.../.claude/projects/...)` was a MIS-diagnosis — that path
  carried a `Read(//…/.claude/projects/**)` deny, and Read/Edit denies extend to recognized
  Bash file commands (`cat`/`head`/`tail`/`sed`) per the doc's deny-reach clause, so it was
  BLOCKED under the round-2 profile, not auto-approved. The REAL round-2 Bash-cat exposure
  was UN-named credentials-class paths allow-scoped-but-not-Read-denied (a stray
  `~/.aws/credentials`, a project `.env` not under `~/.claude`) — `Bash(cat
  ~/.aws/credentials)` would have auto-run, composable with the WebFetch(arxiv) egress. The
  repair is the doc's own stated
  remedy (round 3, now BELT): an enumerated deny rule per carve-out command ("to require a
  prompt for one of these commands, add an `ask` or `deny` rule for it"; deny at any level
  can never be overridden — deny supremacy is doc-verified), leaving the session's read
  surface exactly the native, allow-scoped Read/Grep/Glob tools. **Round 4 supersedes
  enumeration as the PRIMARY close with a bare `Bash` deny (R4-3, next bullet)** — because
  the doc set is non-exhaustive ("include[s]") and the git member escaped, enumeration alone
  could never be load-bearing. The four `Bash(git ...)` allow
  rules are RETAINED but re-labeled a second time (R5-1): under the shipped bare-`Bash`-deny
  profile they are DEAD RULES — a bare tool-name deny removes the tool from context
  entirely and deny supremacy means no allow rule can resurrect it, so neither these rules
  NOR the built-in read-only-git carve-out fires at all (the round-3/round-4 comment "the
  carve-out auto-approves read-only git regardless" was true of the PRE-bare-deny profile
  and is REFUTED for the shipped one — corrected at both sites, here and in the JSON
  block). They stay in the file solely as declared intent for rung-3/4 rebuilds that lack
  the bare deny, where the carve-out does bind. The
  allowlist-inversion claim survives for tools; for Bash the true shape is
  allowlist-plus-enumerated-carve-out-denies, stated plainly. Re-verified live at the
  permissions doc, blue round 3 (the carve-out set quoted above is the doc's, which is
  BROADER than the round-3 gap summary — nine additional commands, all now in the deny
  enumeration).[^PermissionsDoc][^RedPatterns] The deny list is belt-and-suspenders on top.
- **Round-4 structural close — the whole Bash carve-out class, not command-by-command
  (R4-3, lead-endorsed cross-cutting direction).** The round-3 repair enumerated the 14
  doc-named commands, but the doc states the set with "include[s]" — NON-EXHAUSTIVE by its
  own phrasing (the same page names `sort`/`sed` as classifier-reasoned commands outside
  the 14), so an unlisted member (`sort`/`file`/`readlink`/`strings`/`less`) auto-ran
  un-denied under the round-3 profile, and the belt Read-denies bound only the NAMED
  `~/.claude` targets, leaving un-named box secrets (`~/.ssh`, a stray `.env`) exposed on
  that channel. The structural fix is a **bare `Bash` deny** in the sleeper profile: a bare
  tool name removes the tool entirely (doc-verified), and §2.2's session steps never invoke
  Bash — git and node run WRAPPER-side — so denying Bash costs the session nothing and
  closes the enumerated members, the unlisted members, read-only git, AND every git-native
  write gadget (R4-2/R4-15) in one rule. The enumerated per-command denies, the git-output
  belt denies (now extended to the `-o`/`-O` short forms and the
  `format-patch`/`archive`/`bundle`/`config`/`gc`/`repack`/`maintenance` writer family —
  R4-2/R4-15), and the hook's git matcher are RETAINED as belt for the rungs that rebuild
  without the bare deny (3–4) and against profile drift. The bare-vs-`*` convention is
  stated in the shipped file: `"Bash"` removes the tool; `"Bash(cmd *)"` scopes a single
  command form. OQ18(c) is extended to a member-enumeration probe under a bare `dontAsk`
  profile (any command that still auto-runs gets its own belt deny). The redundancy is
  deliberate and additive — the report never subtracts a guard to satisfy the structural
  one.
- **Git channel — inverted to a read-ALLOWLIST at the hook (R4-2, lead-endorsed).** Where
  Bash IS reachable (**a rebuilt rung — 3–4 — or profile drift; "the Workflow seat agents"
  are REMOVED from this list round 5, R5-1: the round-4 bullet contradicted §4.3 layer 4
  (iii), which makes the seats full permission-engine subjects INSIDE the sleeper boundary
  — the horn is now picked (seats BOUND, layer 4), so Bash is NOT reachable for the
  nightly seat population by design, and the R4-2 premise as printed was false for
  them**), the round-3
  git posture — pin the allowed read forms, deny known write flags — was still a denylist
  and its siblings escaped: `git format-patch -1 -o /tmp/probe` exits 0 and writes an
  arbitrary out-of-repo patch (leaf-verified this box, R4-2), `-o` matched none of the three
  long-form `--output` denies, and `git archive -o`/`git bundle create`/`git config
  --global` are further writers. The sleeper-guard hook (which already parses Bash) is
  therefore INVERTED on the git channel per invariant 6: it DENIES any `git` argv that is
  not one of the exact allowed read forms (`git status`, `git diff`, `git log` in its pinned
  forms), so a new git writer defaults DENIED rather than needing a new belt line. The
  subcommand boundary — `config`/`gc`/`repack`/`maintenance` write `.git/config` or
  `~/.gitconfig` (a pager/alias write is arbitrary command exec on the human's next
  interactive git use) — is inside the allowlist's deny-by-default and named in OQ18(c) as
  its standing test. This is the enforcement of record for the git channel; the belt denies
  above are defense-in-depth.
- **The one write gadget found inside "read-only" — closed (R3-15):** `git log
  --output=<path>` exits 0 and creates the file — an arbitrary-file write via a
  git-native flag, with NO permission-engine bug involved (reproduced independently on
  this box by red round 3 and blue round 3; blue's reproduction ran without any prompt.
  **Attribution re-scoped round 4 (R4-11):** both reproductions ran under
  `defaultMode: "auto"` (direct read of `~/.claude/settings.json`, no Bash/git allow
  rules), where the AUTO classifier is the approving layer — so "no prompt" is CONSISTENT
  WITH carve-out classification but does NOT isolate it; the isolating dontAsk-zero-allow
  probe was denied twice from lens seats by the seat's own auto-mode classifier (red
  friction, round 4) and is deferred to build-time (OQ23). The safety conclusion is
  UNAFFECTED — under the round-4 bare `Bash` deny the gadget is closed regardless of which
  layer would have classified it, and the hook matcher is chosen-conservative, not
  proven-necessary).
  Three-part close: the git allow rules are pinned to exact argv (no wildcards); belt
  deny rules match `* --output=*` / `* --output *` / `* --output-directory*` in any
  command string (deny beats the carve-out, doc-stated); and the enforcement of record is
  the sleeper-guard hook's Bash-write matcher, extended to git-output-class flags
  (`--output`, `--output-directory`, `-O`) — a blocking hook stops the call before
  permission rules are evaluated, carve-out included.[^PermissionsDoc] An out-of-repo
  `--output` target is exactly the R3-8 snapshot blind spot, which is why §6 row 4 and
  §4.3 layer 5 cite "no remaining write gadget" only AFTER this fix (sequenced by
  construction). OQ18's test matrix now includes git-native output flags and
  redirection-under-carve-out.
- **Deny-precedence physics, verified against the permissions doc:** deny > ask > allow;
  "If a tool is denied at any level, no other level can allow it"; deny rules survive even
  `--allowedTools` flags; command-line arguments sit BELOW managed settings and cannot
  override any deny.[^PermissionsDoc]
- **Drafting details a naive spec would get wrong:** path rules must be written as
  `Edit(path)` — `Write(path)` rules are "accepted but never matched" (startup warning
  since 2.1.210; Edit rules cover all file-editing tools); a `Read` deny also blocks Edit on
  the same path from v2.1.208 [minority: lane-2/primary-literature — the Read-deny detail];
  Windows paths normalize to `/c/...` form for absolute-anchored rules [minority:
  lane-3/local-probe].[^PermissionsDoc]
- `disableBypassPermissionsMode: "disable"` — "A user can set it in their own settings to
  lock themselves out of bypass mode" — closes the escalate-to-bypassPermissions route
  [minority: lane-3/local-probe].[^PermissionsDoc] Placement matters (R1-8): the key lives
  INSIDE the `permissions` object; the round-0 sample put it at top level, where it is
  silently ignored — the lockout would never have engaged.
- **Anchoring (added round 1, R1-12):** bare `path` rules anchor at the process current
  directory; `/path` anchors at the settings file's location; `//path` at the filesystem
  root.[^PermissionsDoc] Guardrail rules in this file are therefore written `//`-absolute,
  AND the wrapper `cd`s to the repo root before exec — stated in scheduling.md as a
  precondition beside the trust-dialog one (Task Scheduler's default start-in directory is
  System32; a forgotten `cd` must not be able to relocate the deny set).

### 4.3 The layered gate stack (each layer names how the loop would route around it, and what stops that)

| Layer | Mechanism | Route-around and answer |
|---|---|---|
| 1. Allowlist-inverted permission profile (§4.2) | `dontAsk` + narrow allows; deny set operator-owned, outside the repo | Loop edits the settings file to widen its own allow list → the file is outside the workspace (not reachable without `--add-dir`), `Edit(**/settings*.json)` + the absolute-anchored `~/.claude` deny cover it, and deny-before-allow means its own allow rules can't help. Known enforcement bugs (§4.1) → this layer is a belt, never load-bearing alone [lane-1's grading] |
| 2. Path-fence hook | `sleeper-guard` PreToolUse (Go, sc-quality-gate class — hooks verifiably fire on subagent writes[^IdeasCorpus]) blocks Edit/Write/NotebookEdit/Bash-write targets outside research/+ideas/ when the sleeper marker identifies the run (env var set by the wrapper + `agent_id`/`agent_type` fields[^HooksDoc]); fence LIVENESS is proven per run by the wrapper's step-0 canary drive, round-2 mechanism (R2-1/R2-2 — the round-1 deny-outcome check could not isolate this layer even in principle: layers 1 and 2 fence the IDENTICAL boundary, so every canary write is denied by layer 1 whenever it is live, and on marker loss — the modal drift case — the fence is dormant, the canary would PASS, and the run would proceed under exactly H4's refuted permissions-only configuration while stamping "enforcement proven"). The repaired mechanism, per invariant 7: the guard, on every sleeper-fence DENY it issues, appends a fired-record line (tool, target path, nonce echo, decision) to the operator-owned dir — hooks are harness-side processes, not model tool calls, so the record is reachable by the guard and unforgeable by the session (Edit denied; Workflow scriptPath pinned; and the Bash surface — round 4, R4-3: the bare `Bash` tool is DENIED in the sleeper profile (the session never invokes Bash), so the entire carve-out class is structurally removed, with the enumerated denies + git read-allowlist hook (§4.2) as belt for rebuilt rungs; forging the record requires a WRITE the session has no Bash channel for). The wrapper's phase-0 canary (nonce-named target — §2.2 step 0) then requires BOTH (a) the write was denied in the event stream AND (b) the fired-record contains the nonce with decision=deny — proving THIS layer fired, not merely layer 1. Marker loss ⇒ the guard never engages ⇒ no fired-record ⇒ ABORT, even though layer 1 still denies the write: fail-closed restored, and OQ2's fence-dormant acceptance test now passes as designed. (Belt, verify-at-build: if stream-json deny events carry distinguishable PROVENANCE — hook-block message vs permission-rule denial — the wrapper parses it as a second witness; OQ21.) Marker-loss base rate (scheduler misconfig, shell nesting, rung migration) still exceeds any adversarial term in §6 (R1-28). Deny-set **compiled into the binary**, not read from editable config [minority: lane-1/adversarial — compiled-in deny-set]. Hook semantics favorable: "A blocking hook also takes precedence over allow rules," "Hook decisions don't bypass permission rules" — layers 1+2 can only tighten each other, never loosen[^HooksDoc][^PermissionsDoc] | Loop modifies hook source in `plugins/` in-repo → the repo copy is not the executing copy — now PINNED, not assumed (R1-15): non-bare runs execute the plugin cache under `~/.claude/plugins/`; the `--bare` recipe's `--plugin-dir` points at the operator-owned read-only copy (`~/.claude/sleeper/plugins/`), NEVER the working tree (round 0 left this unreconciled — under the preferred recipe as written, the repo copy could have BEEN the executing copy, voiding this layer's independence leg). `Edit(plugins/**)` is denied anyway, and a modified repo copy reaches either executing copy only through the human-gated plugin-update path |
| 3. Self-protection of the guardrail definitions | Every file that *defines* layers 1–5 (permission profile, plugin hooks, scheduled-task definition, the command markdown of /self-improve and /graduate, the guard binary) is outside the write boundary; the hook additionally denies writes to the plugin cache, settings, hooks.json, and its own binary path. The empty-bin bootstrap guard keeps hook failure loud[^HooksJson][^Backlog], and the wrapper aborts if the guard binary is missing (fail-closed, not no-op) [minority: lane-1/adversarial — the fail-closed abort]. The loop can propose changes to its own guardrails only as **idea stubs** walking the same human-gated graduation path — closing the DGM failure mode by construction rather than instruction[^DGMSakana]. **Round-3 extension (R3-9 — the reflexivity blindspot beneath the invariant-7 watch): the protected enumeration now names the WATCHMEN themselves** — `sleeper-wrapper.mjs`, `harvest.mjs`, `sleeper-permissions.json`, `sleeper-mcp.json`, all living in the operator-owned dir (`~/.claude/sleeper/`), invoked by the scheduled task via ABSOLUTE PATH from that copy, never from the repo working tree; the snapshot guardrail hash set includes them. The self-verification limit is OWNED, not hidden: the wrapper is actor, observer, and abort authority for all five gate controls, and no in-run mechanism can verify the verifier — the R2-8 hash check runs BY the wrapper, so a compromised wrapper attests nothing. Its integrity rests on the operator-owned dir sitting outside every session write surface (§4.2 denies + dontAsk + the R3-14 carve-out denies) plus OS file permissions; the residual — out-of-band compromise of the operator-owned dir — is a box-compromise scenario, not a loop-autonomy scenario, and is stated rather than papered. Cheap liveness in place of impossible self-attestation: the SessionStart line always prints its last-check timestamp (§3.4 — silent non-fire is thereby visible), and the doctor line prints the wrapper-file hashes' last-change dates | Loop disables hooks by editing settings → settings paths are inside the deny-set; residually, chmod-readonly on guard files as defense-in-depth — a DESIGN-PROPOSED measure of this report's own, not community-sourced (corrected round 1, R1-13: the #22055 thread documents the PreToolUse protected-files workaround verbatim; chmod-444 is not in the thread) [minority: lane-1/adversarial][^PermAskBypass] |
| 4. Subprocess residual (the honest gap) | File rules "don't apply to arbitrary subprocesses that read or write files indirectly, like a Python or Node script that opens files itself"; OS-level enforcement requires the sandbox.[^PermissionsDoc] Under `dontAsk` + the round-4 bare `Bash` deny (R4-3) the Bash channel is closed at the tool boundary for the sleeper session, but the Workflow-tool script is itself a subprocess | In order of preference: (a) enable Claude Code **sandboxing** for the scheduled profile so filesystem writes outside the run tree are OS-blocked for Bash and children (permissions + sandbox are the documented defense-in-depth pair); (b) where the sandbox is unavailable (current Windows box), the round-1 shape (R1-16) removes all MODEL-AUTHORED script execution from the session — re-scoped round 2 (R2-3: the round-1 "removes ALL script execution" was falsified by the design's own step 4, which invokes FEOV nightly, and the shipped command's execution locus is MIXED per probe[^ResearchCommand]: setup/capture are session-Bash node calls, the debate engine is a **Workflow tool** invocation). The true premise, stated: (i) the two Bash node scripts are re-hosted WRAPPER-SIDE (§2.2 steps 0/4/7), so the in-session Bash surface is **closed structurally round 4 (R4-3): the bare `Bash` tool is DENIED in the sleeper profile** — a bare tool name removes the tool entirely (doc-verified), the session's steps never invoke Bash, so the dontAsk read-only carve-out CLASS (enumerated members, the doc-non-exhaustive unlisted members, read-only git, and every git-native write gadget) is removed at the tool boundary rather than command by command. This supersedes the round-3 enumerate-and-pin posture as the PRIMARY close; that posture is retained as belt for rungs that rebuild without the bare deny (3–4) and against profile drift, and for those environments the git channel is inverted to a read-ALLOWLIST at the hook (R4-2: `git format-patch -1 -o <path>` → exit 0 arbitrary out-of-repo patch was a git sibling the round-3 `--output` denylist missed; the hook now denies any git argv not an exact allowed read form). With Bash denied, no surviving rule OR carve-out entry grants the model argv that chooses a subprocess write target; OQ18's extended matrix (redirection, compound forms, git-native write flags, unlisted carve-out members, git write subcommands) is the standing belt-verification test, not the load-bearing close (the round-0 `Bash(node scripts/*)` grant contradicted "read-only analysis tools" in-document AND mismatched the printed layout); (ii) the ONE remaining in-session script channel is the Workflow tool running debate.js — PINNED CODE from the read-only operator-owned plugin copy, model-influenced only through schema-bound args (topic/runDir/lanes/models), with the `Workflow` tool added to the §4.2 allow list and the sleeper-guard hook additionally matching Workflow calls and denying any scriptPath outside the plugin copy (PreToolUse fires on every tool; whether permission RULES can scope Workflow's scriptPath natively is OQ20); (iii) **the seat-population horn, PICKED and honestly evidenced (round 5, R5-1 — the round-4 text both asserted the seats were full permission-engine subjects AND listed them as a surface where Bash was reachable; both cannot hold): the workflow's SEAT AGENTS are INSIDE the sleeper boundary — bound by the bare `Bash` deny — as a DESIGN REQUIREMENT.** The evidence honestly scoped: what is leaf-verified is layer-2 HOOK coverage from INTERACTIVE runs (sc-quality-gate fired on workflow-agent writes; red-auditor wrote its memory under hook coverage[^IdeasCorpus]) — that does NOT carry layer-1 `--settings` inheritance for debate.js-spawned seats in a headless run, which is UNPROBED and is now a named acceptance leg, OQ23(e). Two consequences derived rather than assumed: (a) closure on the seat surface never rests on the inheritance assumption alone — the sleeper-guard hook is GENERALIZED per invariant 6 to deny the WHOLE Bash channel for sleeper-marked runs (the hook already parses Bash; seats need no Bash — §2.2 step 4 owns the capability cost), so even if OQ23(e) finds seats outside the `--settings` profile, the hook layer (whose seat-coverage IS the leaf-verified part) closes Bash there, and R1-13 read+egress does not re-open on the seat surface; (b) if OQ23(e) finds seats unbound AND the headless hook leg (OQ2) also failed, the run never starts — the phase-0 canary aborts on a missing fired-record, which is the fail-closed backstop for both probes at once; (iv) the workflow SCRIPT's own Node fs writes are a layer-4-class subprocess residual — bounded HONESTLY round 3 (R3-8): the snapshot compare detects IN-REPO working-tree deltas and guardrail-file tampering; an OUT-OF-REPO subprocess write (`~/.claude`, a sibling tree, `/tmp`, exfil-to-disk) is invisible to both porcelain (repo-scoped) and the guardrail hash set (named files), so the out-of-repo residual is bounded NOT by detection but by pinned code + the absence of a write gadget in the granted surface — citable only now that the R3-15 git-output gadget is closed (the two fixes are sequenced by construction; guardrail files themselves ARE hashed, so that named case holds either way). Compound-command/redirection/traversal matching for the surviving git rules is a named build-PR test (OQ18); plus a hook clause blocking Bash whose command text targets the guardrail set ("a blocking hook … stops the tool call before permission rules are evaluated"[^PermissionsDoc]); and layer 5's tamper evidence is wrapper-snapshot-backed (R1-26), no longer working-tree-only. Residual RISK-ACCEPTED (§6 row 4, re-argued round 1 WITHOUT the benign-actor premise — R1-21) and stated honestly in docs/scheduling.md rather than papered over |
| 5. Git as the promotion gate (the load-bearing layer) | The loop never commits and never pushes (`git commit`/`git push` denied — lane-2 denies both; lane-3 denies push and notes sc-push-freeze-guard warns-only by design, so the loop profile upgrades it to a hard deny for loop sessions[^PushGuard]). The loop's entire output is an uncommitted working-tree delta under research/+ideas/ that the operator reviews next session — `git status`/`git diff` **is** the review surface, and anything outside the boundary shows up as an anomaly [minority: lane-2/primary-literature — the uncommitted-delta review model]. Round-1 hardening (R1-26): an uncommitted delta is SELF-ERASING for exactly the residual it backstops — the layer-4 subprocess channel that can write out-of-boundary can also revert before review, and run N+1 fires before run N is reviewed; DGM's lineage was durable and externally inspectable, and an uncommitted working tree is its opposite. So the wrapper takes an out-of-session snapshot (`git status --porcelain` + SHA-256 hashes of the guardrail file set) at run START and END, appended to an operator-owned log outside the repo: a write-then-revert inside the run shows as a start/end delta the model cannot reach. **The delta now has a WATCHMAN (round 2, R2-7 — a window nobody reads is telemetry, not detection; round 1 reintroduced for tamper evidence the exact undefined-receiver defect R1-29 had fixed for loop death), and the watchman covers EVERY exit, not only the happy path (round 3, R3-7):** the wrapper snapshots at every exit path it observes — step 7, deterministic aborts, budget/turn kills, canary aborts, its own crash handler, and DEAD-marking time — because the compare ran ONLY at step 7 in round 2, a run dying earlier took no end snapshot, the next run's "start-vs-prior-END" had no referent, and aborts are model-influenceable, so the write-then-force-abort sequence was an unbacked hole in this leg. The compare is end-vs-start plus start-vs-THE-LAST-RECORDED-SNAPSHOT-OF-ANY-KIND; a missing chain link fails the next preflight closed (invariant 7). On any mismatch it writes a dated TAMPER flag beside the ledger, the doctor line prints it, and the NEXT preflight fails closed until the human clears the flag — so run N+1 cannot fire past unexamined tamper evidence, and detection latency is one wrapper-observed EXIT (abort paths now included), not one human review cadence. **Unobserved-exit case owned (R4-6):** a wrapper hard-kill leaves no end snapshot, but the compare chains to the last-recorded snapshot (the START snapshot), so the referent survives; the run-window is closed retroactively at DEAD-mark time and read as `retroactive-uncertain` with its markerless sweep confined to sleeper date-key naming (§1.5), and the resume path re-opens the dir under a fresh observed window — so an unobserved exit degrades to snapshot-chain + resume, never to a silent sweep of human-present dirs. **Sensing scope stated honestly (R3-8):** the snapshot senses in-repo working-tree deltas and the guardrail hash set; it does NOT see arbitrary out-of-repo subprocess writes — that residual is bounded at layer 4 (iv) by pinned code + no-remaining-write-gadget (post-R3-15), not by this detection. This restores the durable tamper-evidence property the DGM analogy actually requires. (Heavier alternative if snapshot review proves insufficient: wrapper auto-commit of the run delta to a `sleeper/` branch — commit denial binds the model, not the wrapper.) Optionally the run operates on a `sleeper/`-prefixed branch or dedicated worktree. **Server-side out-of-band enforcement** [minority: lane-1/adversarial]: promotion into rules/skills happens ONLY via PR into main under branch protection + CODEOWNERS on `plugins/**` and `.claude/**` — controls no local write can alter. Cloud routines get this natively: pushes restricted to `claude/`-prefixed branches by default[^RoutinesDocs] | The loop cannot route around GitHub's server-side enforcement from a local session; the erosion path left is *social* (human rubber-stamping PRs), which is a review-culture problem no mechanism solves — named honestly in scheduling.md [minority: lane-1/adversarial]. A purely local repo downgrades to layers 1–4 + review discipline — accepted because the suite's operating model already assumes a GitHub-hosted repo with PR review (PR #1 flow) |
| 6. Invocation gate on /graduate | `disable-model-invocation: true` — the harness withholds such skills from scheduled/self invocation; fires reach Claude as plain text[^ScheduledTasks] [minority: lane-1/adversarial — the mechanism]; graduation exists only as a human keystroke — doc-verified for SCHEDULED fires; the in-session hostile-invocation case is exactly open question 3, so the strong form is qualified here where builders read, not only in §8 (R1-24) | A prompt-injected "run /graduate" in harvested friction text lands as inert text for scheduled fires; for the in-session case (OQ3) layers 2+5 hold even if invocation slipped |
| 7. Process (beyond mechanism's reach) | The last gate is the human reading the stub's provenance chain at /graduate; the stub contract makes provenance mandatory and checkable | A poisoned stub persuading the human is out of mechanism's reach [minority: lane-3/local-probe — naming it]; mitigated by the provenance contract + red's adversarial pass at graduation |

**Semantic-consent continuity** [minority: lane-3/local-probe]: the shipped skill already
defines the loop's posture at a wall — "AFTER hitting a tool limit, YOU MUST escalate with a
recommendation — YOU MUST NOT attempt an autonomous hot-fix. As a subagent, the escalation
channel is your return envelope: report the gap as friction."[^SemanticConsent] The nightly
run's friction lands in the run record; the loop eats its own complaints the next morning.
The consent doctrine and the input pipeline are the same mechanism seen from two sides.
Round-1 caveat (R1-25): "eats its own complaints" is precisely the authorship loop §1.5
now bounds — sleeper-origin friction is tagged and recurrence-capped at 1 at harvest, so
the loop's own nightly complaints inform the docket without being able to ratchet their
own ranking.

**Verdict on H4: SUPPORTED — and the permission-rules-only variant is REFUTED at the leaf
node.** Confidence HIGH on permission/hook semantics (verbatim doc + live probe); the
subprocess residual is honestly open on this box.

---

## 5. H5 — Cost discipline for unattended runs (SUPPORTED; quota-introspection falsifier PARTIALLY CONFIRMED)

### 5.1 Hard ceilings the harness actually provides (primary-doc verified)

- **`--max-budget-usd`**: "Maximum dollar amount to spend on API calls before stopping
  (print mode only)."[^CliReference] Exceed-behavior is under-documented (no exit-code
  contract published) — the build PR must verify by forced-tiny-budget test and record the
  observed behavior [minority: lane-3/local-probe — the under-documentation catch].
- **`--max-turns`**: "Limit the number of agentic turns (print mode only). Exits with an
  error when the limit is reached" — the documented-error backstop.[^CliReference]
- **`--model` / `--fallback-model`**: pin the bulk tier explicitly; the fallback keeps a
  retired/overloaded model from silently failing the night's run. (Round-1 correction
  R1-10: the current cli-reference carries no print-only marker on `--fallback-model`, and
  documents a persistent `fallbackModel` setting besides — the round-0 label overstated
  what the "quoted verbatim" footnote carries.)[^CliReference]
- **`--output-format json` → `total_cost_usd`** per invocation, feeding the loop's own
  cost.md class without transcript parsing — live-verified (probe P1).[^HeadlessProbe][^HeadlessDocs]
- **No harness-native cross-run (monthly) budget exists for a local scheduler.** For org API
  accounts, the Admin API offers `/v1/organizations/usage_report/messages` and
  `/v1/organizations/cost_report` (≈5-min freshness, 1/min polling) — but requires an Admin
  key, not available to subscription-auth (claude.ai) usage, which is exactly the auth mode
  routines require [minority: lane-1/adversarial — the Admin-API trace].[^UsageAPI][^RoutinesDocs]
  Console monthly spend limits exist at organization and workspace level — settable in
  the **UI only: no API endpoint reads or sets SPEND limits. RATE limits, by contrast,
  ARE API-readable** (requalified round 1, R1-9 — the round-0 flat "no endpoint to read"
  was stale): the Rate Limits API (`/v1/organizations/rate_limits` + the per-workspace
  variant) reads configured org and workspace rate limits — read-only, Admin-key-only
  [^RateLimitsAPI]. The design conclusion survives on auth grounds: an Admin key is
  unavailable to subscription-auth (claude.ai) usage — exactly the auth mode routines
  require — so there is still nothing a subscription-auth scheduler can poll pre-launch
  [minority: lane-2/primary-literature — the Console-limits trace].[^ConsoleLimits] H5's
  falsifier is thereby partially confirmed, and the design says so honestly.
- **The honest degraded guard — a local ledger (relocated round 1, R1-19):** the wrapper
  appends each run's `total_cost_usd` to `~/.claude/sleeper/ledger.jsonl` — the
  operator-owned dir hosting sleeper-permissions.json, WRITTEN BY THE WRAPPER outside the
  model session and outside the loop's write surface. (The round-0 placement,
  `ideas/.sleeper-ledger.jsonl`, violated invariant 4 for the design's only cross-run
  control: "append-only" was unenforced under `Edit(ideas/**)`, git tamper evidence
  covers only committed states the loop cannot commit, and a mangled JSONL line would
  have silently disabled the guard — the report's own STOP/row-7 evidence says a
  rewritable budget file eventually gets rewritten.) Preflight sums month-to-date and
  **skips the run** with a dated skip record when over cap, and **fails closed on an
  unparseable ledger** (skip, don't guess). The invariant-4 layer covering the monthly
  guard is therefore the wrapper itself; the skip record's defined reader (R1-29) is
  doctor's "last successful sleeper run" line. Self-accounting inherits the attestation
  ceiling (in-run checks catch shape, post-hoc audit catches
  vacuity[^EfficiencyPlan]) — but the number comes from the harness's own cost meter, not
  the model's self-report, the strongest primitive available locally. Post-hoc audit
  stays cost-audit.mjs's job. Plus a standing monthly workspace cap set once in the
  Console UI where the API-key path is used.
- **Auth-path caveat** [minority: lane-2/primary-literature]: subscription (OAuth) sessions
  bill against plan usage limits — run 4 died mid-flight on exactly that wall with no
  warning; API-key runs bill per token against Console limits. The scheduled recipe must
  name which path it uses; `--bare` forces the API-key path anyway.[^HeadlessDocs][^CostRecord]
- **Cloud routines** meter against subscription usage with a daily routine-run cap and
  optional metered overage; absent usage credits, runs are "rejected until the window
  resets" — a platform-side backstop for rung 3, not a substitute for the wrapper's per-run
  ceiling.[^RoutinesDocs]

### 5.2 Tiering: cheapen redundancy and mechanics, never judgment

| Stage | Tier | Ceiling / rationale |
|---|---|---|
| preflight, harvest, score, record | zero tokens (scripts) | n/a — pure mechanics |
| pick + stub drafting | bulk (haiku/sonnet class) | `--max-turns` small; single agent; worst case is one wasted bounded run |
| bounded research pass (daily) | bulk-tier lanes; judgment inherit-session NOT granted headless — pin judgment to bulk in daily runs and accept the grade; **seats run BASH-FREE (R5-1): smoke-lane + citation-pass work is Read/Grep/Glob + qmd + WebFetch(arxiv), repo state wrapper-staged — command-execution leaf probes are a graduation-run capability, priced in §2.2 step 4** | smoke-scale (~50k tokens measured)[^Backlog]; the daily run produces *stubs*, not verdicts; its output is always re-judged at graduation |
| whole daily run | — | `--max-budget-usd` **$2–5** (per-run hard CEILING — an anomaly bound); **expected** per-run spend ~$0.10–0.50 at bulk-tier list rates (the ~50k-token smoke shape[^Backlog]; probe P2's full plugin-command run measured $0.058[^HeadlessProbe]); monthly ledger cap ~$50 (operator-tunable). Arithmetic reconciled round 2 (R2-18): 30 × expected ≈ $3–15/month, so the $50 cap carries ≥3× headroom over expected cadence and binds only under anomaly — cap-trip is an ANOMALY SIGNAL, not the intended month-end throttle; a nightly ceiling-priced run ($2–5 × 30 = $60–150/mo) would by design trip the cap mid-month, which is the cap doing its job, and the resulting skip streak is distinguishable from loop death because skip records carry the REASON, printed on the dead-man line (§3.4) [minority: lane-1/adversarial — the specific dollar defaults] |
| /graduate full debate | normal FEOV routing (bulk + judgment tiers) | human-present; existing FEOV ceilings + stop-and-resume; never cheapened |

Rationale from measurement: a full debate run cost **$414.97 at list rates** (42 agents,
1975 api-turns, cache traffic 99% of tokens, judgment-seat premium cache-RATE-driven); run 3
was $149.95 [minority: lane-3/local-probe — the run-3 figure].[^CostRecord][^EfficiencyPlan]
(R2-22: the run-3 figure's artifact of record is `plans/efficiency-phase.md` §I, not
cost.md — the second marker pins it.) Daily
full-strength is indefensible ($4.5k–12k/month) — three orders of magnitude above a
defensible daily unattended spend; the smoke shape prices in single-digit dollars even at
premium tiers. The daily loop must never spawn full FEOV; the expensive machinery is
reachable only through the human-gated /graduate. The daily loop performs NO final judgment
— its every output is a proposal judged later — so the efficiency doctrine's protected
category (judgment, the adversary, the full re-read) is never exercised unattended, and
therefore never cheapened. Run-4's measured physics says keeping strong models OUT of the
nightly loop is the dominant term — bulk-tier freight "dwarf[s] every
lever."[^EfficiencyPlan]

List-rate reference points [minority: lane-2/primary-literature] (volatile; leaf-verified
against the platform pricing page at red's round-1 audit, upgrading round 0's
aggregator-carried MEDIUM — R1-11): Haiku 4.5 $1/$5 per MTok in/out; Sonnet 4.5/4.6 $3/$15
(Sonnet 5 intro $2/$10, then $3/$15 from 2026-09-01); **two frontier tiers, named
(R1-11):** Opus 4.5–4.8 $5/$25 and Fable/Mythos 5 $10/$50 — the round-0 "frontier ~$10/$50"
was true only of the Fable/Mythos class. Cache reads 0.1×; prompt caching cuts cached
input ~90%. The new-generation tokenizer (Opus 4.7 and later Opus models, Fable 5,
Mythos 5/Preview, Sonnet 5 — completeness fix round 3, R3-17: the round-2 list omitted
the Opus 4.7+ members) counts ~+30% more tokens than legacy
counting, so cross-era dollar comparisons (including this report's $414.97/$149.95
anchors) are approximate, not exact. The Batch API is a documented flat 50% off; the ≤24h
async-window sub-claim, carried MEDIUM at round 1, is resolved HIGH round 2 by red's
batch-processing-page fetch ("Batches expire if processing does not complete within 24
hours"). Batch as a lever is demoted to a FUTURE note (R1-23): the nightly
loop runs through the claude CLI, which offers no Batch routing today — a nightly loop
would be the ideal Batch customer (nobody is waiting) if and when a routing mechanism
exists, and not before.[^Pricing]

### 5.3 Stop conditions yield honest partials

Run 4's death at the monthly spend limit is the type specimen: null-guard abort, blackboard
state intact, resumable cached session, honest UNVERIFIED assembly — losing no paid-for
work.[^FrictionRun4][^CostRecord] Sleeper inherits: every stage writes the blackboard as it
goes (scored table after step 2, pick after step 3, partial research artifacts during step
4, stub file before polish); `--max-budget-usd`/`--max-turns` exhaustion exits nonzero; the
scheduler records the session id; the next scheduled fire **resumes** rather than restarts
(§3.4 idempotence; resume with original models — cache-safe, per the standing
stop-and-resume practice[^ResearchCommand]). An aborted night leaves a resumable stub plus a
dated abort record, not nothing.

Termination stays judged where judgment exists (graduation runs: telemetry +
stop-and-resume, ratified run 4; automatic severity-floor termination was evaluated and
REJECTED) — the daily loop's ceilings are pure cost caps, not quality judgments, which is
why they may be automatic [minority: lane-1/adversarial — the explicit
cost-cap-vs-judgment distinction].[^EffReport]

**Cost telemetry feeds back** [minority: lane-3/local-probe]: the ledger and per-run cost.md
are harvest.mjs inputs — harvest runs wrapper-side (R1-16) and stages a read-only copy of
the operator-owned ledger into the run inputs, so the loop consumes its own spend record
without holding any path to the enforcement copy (R1-19) — and can propose its own diet
(but never enact one: a model-tier change is a settings/plugin edit, outside the write
surface).

**Verdict on H5: SUPPORTED, with the monthly guard honestly degraded to
ledger-plus-static-ceiling** (the falsifier's fallback, designed-in rather than discovered
later). Confidence HIGH on flags/docs (verbatim fetches + live probe); MEDIUM on the ledger
design (unbuilt); pricing figures leaf-verified round 1 (R1-11). Quota introspection,
requalified (R1-9): CONFIRMED-NEGATIVE for spend limits (no API read or set) and for
anything a subscription-auth scheduler can poll; rate limits are API-readable but
read-only and Admin-key-org-only — outside this scheduler's reach.

---

## 6. Risk matrix (merged; likelihood × impact × complexity-to-mitigate; risk-accepted rows argued)

| # | Risk | L | I | Cx | Disposition |
|---|---|---|---|---|---|
| 1 | Headless plugin-command expansion regresses (#837 class) or `--bare` becomes the `-p` default and the nightly run silently loses plugins | Low–Medium (doc explicit today; flip announced) | High (loop dead) | Low | Mitigate: step-0 loaded-plugins assertion aborts loudly; the two-legged Phase-4 acceptance test (§3.3, restated round 5 R5-2: wrapper `--manual` produces the run dir; `claude -p "/self-improve"` produces NONE) is the standing verify; version pin; idempotent scheduler makes failure loud-but-harmless |
| 2 | Daily stubs too shallow to graduate (H2 falsifier) | Medium | Medium | Low | Measurable via ledger stub→graduation survival rate; named revisit trigger: switch to weekly-full-strength |
| 3 | Proposal fatigue at the human gate | Medium (Dependabot base rate) [minority: lane-2/primary-literature] | Medium | Low | 1-stub/run cap + dedupe against open stubs younger than the staleness window; stubs auto-stale at 30 days via harvest's filename-date sweep and re-enter the docket flagged `stub-stale` (mechanism specified §1.4/§2.3 round 1 — R1-22; round 0's "age out visibly" was policy-without-mechanism); cadence revisit trigger in §1.4 (R1-14); `graduation-queued` stubs carry their own M-day `queued-stale` re-surface for human re-confirmation (R3-13); **round-4 root invariant (R4-9): every status has a stated re-surface path — no status is timer-free** — graduated re-enters `regression` on class recurrence, rejected dedupes for a 90-day/rate window then re-surfaces `rejected-recurring`, ending the per-status patch chain R1-22→R2-11→R3-13 with the invariant stated once in §1.4 |
| 4 | Subprocess write outside boundary (layer-4 residual; no full sandbox on Windows) | Low-Medium (re-graded round 2, R2-3: round 1's Low rested on "no script execution" — false as printed; the true surface is pinned git argv PLUS the Workflow channel running pinned debate.js) | High (boundary breach) | Medium | Layered (allowlist + hook + wrapper snapshots); residual RISK-ACCEPTED, re-argued round 2 on the TRUE premise (R2-3, superseding the R1-21 re-argument whose leg (a) the round-1 text falsified) and still WITHOUT actor benignity: (a) channel narrowness, restated round 4 on the TRUE Bash surface (R4-2/R4-3) — the sleeper session's subprocess surface is now closed STRUCTURALLY by the bare `Bash` deny (R4-3: a bare tool name removes the tool entirely; §2.2's steps never invoke Bash), so the round-3 "deny-enumerated per command" claim is CORRECTED — enumeration was never exhaustive (the doc set is stated "include[s]", R4-3) and the git member was retained un-enumerated with sibling writers escaping (`git format-patch -1 -o <path>` → exit 0 arbitrary out-of-repo patch, R4-2 leaf-verified; **R5-5: the escape recurred in the ATTACHED form `-o<path>` INSIDE the round-4 belt extension itself — the belt is declared KNOWN-incomplete on attached short forms and is not extended again; the bare deny + hook allowlist are the closes of record, OQ18(c) leaf-tests the hook's attached-form handling**); those are closed by the bare deny for the session, with the enumerated denies + git read-ALLOWLIST hook (§4.2) as belt for rebuilt rungs (git exception named there, OQ18(c) its standing test). The surviving subprocess is the Workflow tool executing PINNED code from the read-only plugin copy under schema-bound args (hook-enforced scriptPath — §4.3 layer 4; **seat agents are bound by DESIGN REQUIREMENT with the evidence honestly scoped round 5, R5-1: hook-layer seat coverage is leaf-verified from interactive runs[^IdeasCorpus], layer-1 `--settings` inheritance for headless-spawned seats is the named probe OQ23(e), and the hook's Bash-channel deny for sleeper-marked runs closes the seat Bash surface independent of that probe**); so the leg's conjunction now reads: a breach needs a permission-engine bug OR an escape past the bare-Bash-deny/Workflow-scriptPath boundary AND a write path in pinned debate.js code — narrower than round 3 claimed because the Bash class is now closed at the tool boundary, not by a list; (b) detection outside the model's reach — wrapper EVERY-exit snapshot COMPARE with TAMPER flag + fail-closed next preflight (R2-7/R3-7: abort paths included, chain-linked to the last recorded snapshot, missing link fails closed); scope honest (R3-8): in-repo deltas + guardrail hashes only — OUT-OF-REPO subprocess writes are bounded by pinned code + no-remaining-write-gadget (post-R3-15), NOT by this detection; (c) L×I×Cx — full-sandbox complexity on Windows exceeds the bounded likelihood × impact today. No premise about actor intent is used. Revisit triggers: any hook-log blocked guardrail-targeting Bash or Workflow call; any TAMPER flag (now wrapper-raised and preflight-blocking, R2-7) [minority: lane-3/local-probe — the trigger] |
| 5 | No programmatic quota pre-check | Certain (no spend-limit API; rate-limit API unreachable at this auth tier — §5.1/R1-9, cell requalified round 2, R2-13) | Low (static ceilings suffice) | Low | RISK-ACCEPT with §5.1's layered static guards — a Console-scraping workaround would be complexity strictly worse than the risk |
| 6 | Scheduler misses days (sleep/laptop) | High | Trivial | Low | RISK-ACCEPT by design: missed-run-tolerant trigger + idempotence; a skipped day costs nothing |
| 7 | Loop self-modifies its guardrails | Low per-attempt, compounding | Critical | Medium | Layers 1–7 (§4.3); the STOP/DGM/AI-Scientist evidence says treat per-attempt probability as nonzero forever |
| 8 | MCP headless flake (open bugs #76239/#68375; #32191 CLOSED as duplicate, kept as phenomenon history — corrected R1-5) starves qmd recall | Medium | Low | Low | Daemon `/health` preflight + strict-mcp-config + Grep/Read degradation (proven workable) [minority: lane-3/local-probe] |
| 9 | Version-bump empty-bin window: nightly run during the update dance hits a hook crash-storm — EXTENDED round 5 (R5-7): sleeper-service ships its OWN hooks.json whose SessionStart hook fires in every INTERACTIVE session, so an unguarded sleeper hooks.json would crash-storm all interactive work during a sleeper-service cache update, not just the nightly run | Medium | Low | Low | The bound now has TWO named files (R5-7): prosthetic-conscience's hooks.json ships the bootstrap guard[^HooksJson] AND sleeper-service's hooks.json (§0 tree) wraps its SessionStart + PreToolUse commands in the SAME guard — guard coverage for the new file is stated in §0, not assumed, and the empty-bin acceptance check (OQ10) covers both files; scheduler preflight can run doctor check; accept residual noise [minority: lane-3/local-probe] |
| 10 | Self-poisoning of the ranking input — edit AND authorship paths (re-scoped round 1, R1-25; mechanism rebuilt round 2, R2-5/R2-6) | Low-Medium (round 2: the round-1 glob-keyed tag was circumvented by the default nightly control flow — the FEOV sub-run dir escaped it; the causal-provenance marker closes that hole, so the grade holds on the new mechanism, not the broken one) | Medium | Low | Stub-files-only write pattern for the edit path; for the authorship path: wrapper-stamped origin markers on EVERY spawned run dir (sub-runs included) read by harvest — never dir names — with recurrence capped at 1 per class (§1.5, R2-5); corroboration is DECIDED severity-gated (R2-6): infrastructure-failure classes enter flagged `sleeper-only` without corroboration (also independently visible on the doctor/dead-man line), ordinary classes require one non-sleeper occurrence; the grade tracks that decision — residual = flag-reliance for infra classes + corroboration-wait for ordinary sleeper-only phenomena, both owned in §1.5. Round-3 closures re-affirming the grade: the rung-0 void is closed (R3-2 — manual runs pass through the wrapper, so markers and gates hold in the DEFAULT mode); the red-memory mirror surface is provenance-tagged and excluded from the corroboration pool (R3-3); the infra-class tag is wrapper-event-log-only, never friction text (R3-5); the window log survives dead runs (R3-4). Round-4 closures re-affirming the grade: the /self-improve payload is moved out of `commands/` into a `disable-model-invocation` trampoline and the §3.4 containment polarity is corrected (markerless out-of-contract dirs are NON-sleeper and CAN corroborate — the residual is a human's deliberate paste-run under the interactive profile, bounded by the window-log sweep + 1-stub cap; R4-1); red-memory writes are DENIED by design under the sleeper profile (nightly seats do not learn — the R3-3 machinery is belt for drift only, tagging at window-added granularity; R4-4); the unobserved-exit window sweep is confined to sleeper date-key naming so a hard-kill cannot sweep human-present dirs (R4-6 — **confinement re-scoped round 5, R5-4: name-keying CONFINES retroactive-uncertain sweeps only, never assigns origin outside one, and the sweep matches the static `<date>_self-improve/` key plus the RECORDED sub-run paths the wrapper logs at creation, so the auto-tag reaches the sleeper sub-run after a hard-kill without ever pattern-matching a human dir**). Round-5 root fix (R5-3): the corroboration pool is now POSITIVELY defined (§1.5) — a dir corroborates only with affirmative non-sleeper provenance (pre-epoch date or interactive-origin marker), anything unattributable is QUARANTINED counts-for-nothing — dissolving the paste-run (R4-1), unobserved-window (R4-6), and rung-2 markerless (R4-13) per-surface residuals in one predicate; the rung-2 automatic-local-poisoning case red sharpened is thereby closed by mechanism, not adoption instruction |
| 11 | Cloud Routine acts through the operator's identity with all connectors default-included | Medium | Medium | Low | If the cloud alternate is used: trim connectors, keep `claude/`-branch restriction ON; document in scheduling.md [minority: lane-3/local-probe] |
| 12 | Unattended overspend | Low | Medium | Low | Per-run `--max-budget-usd` + `--max-turns` + ledger preflight (operator-owned, fail-closed — R1-19); platform caps on cloud |
| 13 | Exfiltration via read+egress: an injected loop reads box-local secrets and encodes them into its egress (web queries, fetch URLs) — added round 1, R1-17 | Low-Medium | Medium-High | Low | Mitigate: Read/Grep/Glob allow-scoped to the repo (dontAsk auto-denies all other READ-TOOL reads; the Bash read carve-out is closed STRUCTURALLY round 4 by the bare `Bash` deny — R4-3 — so the R1-17 read-scoping holds on the Bash channel too. **Scope stated over the TRUE population round 5 (R5-1): for the top-level session the bare deny is doc-grounded; for the debate.js-spawned SEAT AGENTS the profile binds by DESIGN REQUIREMENT with layer-1 inheritance the named probe OQ23(e) — and the seat-surface Bash close does not wait on that probe, because the sleeper-guard hook now denies the whole Bash channel for sleeper-marked runs per invariant 6, and hook seat-coverage is the leaf-verified layer[^IdeasCorpus]; so read+egress does NOT re-open on the seat surface under either probe outcome**). **Postmortem corrected round 4 (R4-5):** the round-3 example of `Bash(cat //…/.claude/projects/…)` on this row's NAMED transcript target was a mis-diagnosis — that path already carried a `Read(//…/.claude/projects/**)` deny, and Read/Edit denies extend to recognized Bash file commands (`cat`/`head`/`tail`/`sed`) per the doc's deny-reach clause, so it would have been BLOCKED under the round-2 profile, not auto-approved. The real round-2 Bash-cat exposure was UN-named credentials-class paths (a stray `~/.aws/credentials`, a project `.env` not under `~/.claude`) that were allow-scoped-but-not-Read-denied — belt-denied round 4 and closed structurally by the bare `Bash` deny; belt Read-denies on `~/.claude` settings/credentials/session transcripts + `~/.ssh`/`.env` classes; WebSearch removed from the nightly profile. Residual egress = arXiv-domain WebFetch URLs + qmd queries (local) — RISK-ACCEPTED with rationale: single-operator box, repo largely public, the remaining channel is narrow, enumerable, and session-logged |
| 14 | Indirect prompt injection via the retrieval surface (qmd corpus text, arXiv third-party uploads) poisons stub provenance or authors tainted friction — added round 1, R1-18 | Medium | Medium | Low | WebSearch dropped from the nightly default; stub contract labels web-derived claims untrusted-origin and bars them from ranking inputs (§2.3); the R1-25 origin-cap bounds the tainted-friction amplification arm; graduation re-derives every claim under full adversarial review with the human present — a poisoned stub buys one wasted graduation candidate, not a rule change |
| 15 | Persistent loop death goes unnoticed (wedged resume, dead scheduler, silent skip streak) — added round 1, R1-29; re-graded round 2, R2-9/R2-10 | Medium | Low-Medium (residual = detection LATENCY, now stated: bounded by the operator's next interactive session or doctor invocation via the SessionStart warning line at N=7 days; unbounded only if the operator abandons the tool entirely — accepted, single-operator box) | Low | Resume cap k=3 → dead-mark + fresh dir; per-cause signature HALT after M=3 same-signature deaths (R2-10); doctor dead-man line with skip REASON (R2-18) + SessionStart-hook staleness warning in interactive sessions (R2-9 — a passively-received channel independent of the sleeper scheduler); dated skip/abort records in the operator-owned dir |
| 16 | Rung migration silently sheds gates (rungs 3–4 lack the wrapper, settings file, canary) — added round 1, R1-27 | Medium | High | Low | Per-rung gate-survival table in scheduling.md (§3.4); rung-3/4 adoption is a graduation-grade decision requiring its own stub + human gate, never a config toggle |

**Pragmatist scope defense (considered and rejected — complexity exceeds likelihood ×
impact in every case; recorded as risk-accepted, not ignored):** (a) a Console-scraping
quota checker (fragile, no API contract, saved risk Low); (b) cloud-VM isolation as the
daily default (adds an infra dependency and loses the local qmd daemon to close a residual
that layers 1–2 + git evidence already bound on a single-operator box); (c) a bespoke
daemon supervisor for qmd; (d) a quota-introspection service; (e) Windows sandboxing built
bespoke.

---

## 7. Pre-flight self-audit (blue discipline)

- Every load-bearing claim footnoted; access dates on all external footnotes (2026-07-17);
  living-source volatility flagged (Claude Code docs, GitHub issue statuses, routines
  research-preview, pricing aggregators).
- Red gap-pattern checks applied across lanes: **Pattern A** — permission issues #22055 /
  #25621 / #6631 and MCP issues #76239 / #68375 fetched directly with statuses quoted;
  #32191 and #837/#14246 flagged as not-individually-refetched (round-0 status; all three
  leaf-confirmed round 1 — see the Round 1 update below). **Pattern B/E** — the
  alert-fatigue figure carried qualitative-only with an explicit not-leaf-verified label;
  STOP percentages flagged for PDF re-pin; pricing figures graded MEDIUM with canonical
  source named at round 0 (upgraded to leaf-verified HIGH round 1, R1-11 — this bullet's
  lag fixed round 2, R2-14: a reader auditing confidence from §7 alone must not carry a
  stale grade); no number laundered. **Live-source drift** — docs re-fetched at draft time,
  not quoted from memory. **File-type blindspot** — no "X doesn't exist" claim rests on a
  single-scope grep; the "no monthly cap mechanism" claim is scoped to the fetched CLI
  reference + Usage-API + Console docs and labeled as such. **Ephemeral instrument** — the
  lane-3 probe outputs live only in the lane transcript; commands stated for re-derivation;
  cheap fix (re-run + commit outputs) offered if red demands it.
- Confidence self-grades stated per section verdict; minority-report provenance tagged
  throughout per the merge convention.
- Cross-lane conflict declared, not resolved silently (§2.2 step 6).
- **Round 1 update (2026-07-17):** all 30 red gaps addressed; per-gap edits in CHANGELOG
  Round 1 with the propagation greps logged (`40 statused`, `1,558`, `#32191`,
  `print-only`, `sleeper-ledger`, `chmod`, `node scripts`, `~$10/$50`, `no endpoint to
  read`, `ICLR`, `improve themselves the more compute`, `two consecutive runs` — each
  corrected token grepped report-wide in BOTH directions, body→footnote and
  footnote→body, per red's incomplete-repair pattern memory). Banked upgrades claimed
  from red's round-1 notes: STOP figures re-pinned at ar5iv §6.2/Table 2 by three
  independent red lenses (OQ8 resolved; precision note — the with-warning rate was
  insignificantly HIGHER); [^UsageAPI] and [^AIScientist] upgraded MEDIUM → HIGH on red's
  leaf fetches; **[^Pricing] upgraded aggregator-MEDIUM → leaf-verified HIGH (R1-11;
  added to this list round 2 per R2-14)**; issue statuses leaf-confirmed: #837 CLOSED
  COMPLETED, #14246 CLOSED
  DUPLICATE (the supersession story holds), #23707 CLOSED NOT PLANNED, #66395 CLOSED NOT
  PLANNED ([DOCS] class), #22055/#6631/#25621 as reported, #32191 CLOSED DUPLICATE
  (correcting round 0's OPEN claim — R1-5).
- **Round 2 update (2026-07-17):** all 22 round-2 gaps addressed (R2-1..R2-16 carried by
  the lead with owed directions — each direction executed; R2-17..R2-22 new-minted by
  red). Two round-2 probes of record: `claude --help` on the pinned CLI 2.1.212 confirms
  `--input-format stream-json` (the R2-1 two-phase canary drive is buildable at the
  pinned version), and the FEOV execution locus is leaf-determined from the shipped
  command file (setup/capture = session-Bash node; debate engine = Workflow tool —
  R2-3). Banked round-2 upgrades claimed from red's notes: [^WindowsHang] MEDIUM → HIGH
  (body fetched; exact regression span v2.1.161–v2.1.168, fixed v2.1.169);
  [^WebSandbox] and [^MissedRun]-anacron → HIGH; [^Pricing] re-fetched live with zero
  drift AND the ≤24h Batch async-window sub-claim resolved HIGH at the batch-processing
  page ("Batches expire if processing does not complete within 24 hours"); `--json-schema`
  invalid-schema error ≥2.1.205 and the rung-3 ~973MB figure pinned; #76239/#68375
  re-confirmed OPEN 2026-07-17. Propagation greps for round-2 corrections logged in
  CHANGELOG Round 2 (written as a catch-up entry alongside Round 3 — the round-2 revision
  shipped in report.md without its CHANGELOG/debate.md blocks, a process defect the lead
  logged; repaired this round).
- **Round 3 update (2026-07-17):** all 17 round-3 gaps addressed (R3-1..R3-8/R3-10..R3-13
  lead-carried with owed directions — each executed; R3-9 recommend-not-block absorbed in
  its cheap parts with the self-verification residual owned in §4.3 layer 3;
  R3-14..R3-17 first-raise, all absorbed). Two round-3 blue leaf verifications of record,
  per critical-stance before absorbing red's build-altering findings: (1) `git log -1
  --oneline --output=<path>` re-run on this box — exit 0, file created, NO permission
  prompt (independently confirming R3-15; re-scoped round 4, R4-11: the no-prompt result
  ran under `defaultMode: "auto"`, so it is CONSISTENT WITH carve-out classification but
  does NOT isolate it — the isolating dontAsk-zero-allow probe is deferred to build, OQ23,
  and the round-4 bare `Bash` deny closes the gadget regardless of the approving layer);
  (2) the permissions
  doc re-fetched live — the read-only carve-out quoted verbatim ("runs them without a
  permission prompt in every mode ... The set is not configurable; to require a prompt
  for one of these commands, add an `ask` or `deny` rule for it"), with a BROADER command
  set than the round-3 gap summary (adds `ls`, `echo`, `pwd`, `wc`, `which`, `diff`,
  `stat`, `du`, `cd`), all now deny-enumerated (R3-14). Invariant 8 added at the lead's
  direction; the R3-2 `--manual` decision, the R3-7 every-exit snapshot, the R3-11
  signature normalization, and the R3-3 memory-surface tagging are its derived instances.
  Zero grade disputes filed round 3 — every required fix priced trivial-to-low, so
  absorption beat contestation a third time (including R3-17, where red offered
  risk-accept: the one-clause fix was cheaper than the acceptance argument).
- **Round 4 update (2026-07-17; bullet added round 5 — red's sub-trivial note that §7
  lacked one):** all 16 round-4 gaps addressed; the two structural closes (bare `Bash`
  deny R4-3, git hook read-ALLOWLIST R4-2) verified against the live permissions doc;
  R4-5/R4-11 each propagated to a second site found by report-wide grep; cap/HALT
  arithmetic recomputed (R4-10); zero grade disputes a fourth round. Per-gap edits in
  CHANGELOG Round 4 with propagation greps logged.
- **Round 5 update (2026-07-17):** all 10 round-5 gaps addressed (all lead-carried with
  owed directions — each executed; zero grade disputes a fifth round: every required fix
  priced trivial-to-low and the two mediums earned their structural fixes). The two
  invariants of record: the seat-population horn picked with the hook Bash-channel deny
  as the inheritance-independent close (R5-1), and the positive corroboration pool
  (R5-3) promoting OQ24 from deferred to built. Corroboration flags banked from red's
  round-5 notes: ten living-source claim sets re-fetched zero-drift 2026-07-17; the
  cap/HALT arithmetic survives three independent recomputes; [^Pricing] re-verified live
  zero-drift; the attached-form gadget leaf-reproduced twice (adopted into R5-5's
  repair). Note adopted: #68375 now carries a GitHub `stale` label (bot auto-close is a
  live drift risk — content still zero-drift; re-check at build); [^MissedRun]'s Task
  Scheduler settings URL 404s and is re-pointed at the StartWhenAvailable API page.
  Propagation greps in CHANGELOG Round 5.

---

## 8. Open questions carried

1. Does headless plugin-command expansion hold under `--bare` + `--plugin-dir` on the
   current CLI? (Probe P2 answered the non-bare case on 2.1.212; the bare case — the
   future default — is untested. The one load-bearing harness assumption; Phase-4
   acceptance test.)
2. Does the PreToolUse write-fence fire identically under `claude -p --bare` with explicit
   `--settings`? (Interactive + subagent firing is verified; headless-specific firing needs
   the Phase-4 smoke test.) Reframed round 1 (R1-28) and repaired round 2 (R2-2): verify
   HOOK LIVENESS, not deny outcome — round 1's acceptance test was unsatisfiable as
   designed (layers 1 and 2 fence the identical boundary, so a fence-dormant run still
   denies the canary and would have passed). The smoke test now verifies: (a) canary
   nonce appears in the guard's fired-record with decision=deny on a healthy run;
   (b) on a deliberately fence-dormant run (marker withheld), the write is still denied
   by layer 1 but NO fired-record appears and the wrapper ABORTS — the test the repaired
   mechanism passes and the round-1 mechanism could not.
3. `disable-model-invocation: true` is documented for scheduled fires; confirm equivalent
   enforcement when a hostile prompt *inside a -p session* tries to invoke /graduate
   (layers 2+5 hold regardless).
4. Cloud-rung qmd: can the routine setup script install qmd + models (~973MB) within
   environment caching limits, or does rung 3 run recall-degraded (lex-only / no qmd)?
5. Subscription-auth cost telemetry: is `total_cost_usd` populated meaningfully for
   claude.ai-auth runs, and does `--max-budget-usd` count subscription usage or only
   API-key billing? (The doc says "API calls"; run 4's death was a subscription wall.
   Needs one measured headless run on each auth path.)
6. `--max-budget-usd` exceed semantics (exit code? partial result? mid-turn stop?) —
   undocumented; forced-tiny-budget test in the build PR.
7. Sandbox availability/behavior on Windows for the scheduled profile (layer 4's preferred
   close).
8. RESOLVED round 1: STOP circumvention figures re-pinned at ar5iv §6.2/Table 2 by three
   independent red lenses — 0.42% (CI 0.31–0.57%); 0.46% (CI 0.35–0.61%) with warning,
   insignificantly HIGHER (two-proportion z-test); 10,000 sampled improvements; syntactic
   detection. PDF pin remains optional publisher-grade polish.
9. qmd daemon lifecycle ownership: scheduler-owned start-if-absent (lane-3's proposal) vs
   doctor-managed service — decide at build.
10. Does `-p` non-bare reliably load marketplace plugins when the plugin cache is
    mid-update (empty-bin window)? Probe P2 says yes when stable; the update-collision case
    is untested. EXTENDED round 5 (R5-7): the empty-bin acceptance check covers BOTH
    guarded hooks.json files — prosthetic-conscience's AND sleeper-service's (whose
    SessionStart hook fires in every interactive session) — verifying each command's
    bootstrap-guard wrapping against a deliberately emptied bin.
11. Are Desktop scheduled tasks available/stable on Windows Desktop today? (Compare table
    implies yes; not probed.)
12. Routines research-preview churn — re-check at build time whether `/schedule` +
    repo-scoped environments have stabilized.
13. Stub survival-rate instrumentation format (ledger field set) — decide at build.
14. Resolve the §2.2-step-6 conflict: backlog-append (lane-2) vs stub-files-only
    (lane-3) — synthesis proposes a loop-owned generated index; needs a ruling.
15. When `--bare` becomes the `-p` default, does the non-bare recipe silently lose plugins?
    (Wrapper pins the CLI version; re-verify on bump.)
16. Fence-by-default polarity (R1-28): is there a reliable in-hook signal that a session
    is non-interactive/print-mode, so the fence can default ON for headless runs without
    fencing normal interactive work? (The canary closes the fail-open regardless; polarity
    inversion would be belt on top.)
17. `disableAutoMode` (R1-8's sibling escape hatch): leaf-verify the key's existence,
    name, and scope in the current permissions doc before adding it to the profile.
18. Compound-command, redirection, and traversal matching semantics for the surviving
    `Bash(git ...)` allow rules (R1-16c) — named build-PR test. EXTENDED round 3
    (R3-14/R3-15): the matrix adds (a) redirection and compound forms UNDER the read-only
    carve-out (`cat x > y`, `;`/`&&` chains — does the classifier reject them, and does a
    scoped deny rule match the full command string?); (b) git-native write flags
    (`--output`, `--output-directory`, `-O`, and the `-o` SHORT form — R4-2: `git
    format-patch -1 -o <path>` → exit 0 arbitrary out-of-repo patch, missed by the three
    long-form `--output` denies) against the extended hook matcher and belt denies; (c)
    EXTENDED round 4 to the SUBCOMMAND boundary of "read-only forms of git" (R4-15) AND a
    MEMBER-enumeration probe of the carve-out (R4-3), both under a BARE `dontAsk`
    zero-allow profile — the isolating configuration the lens seats could not reach (R4-11,
    deferred here): enumerate (1) which git SUBCOMMANDS classify as read-only — name `git
    config` (writes `.git/config`, or `~/.gitconfig` with `--global`; a pager/alias write
    is arbitrary command exec on the human's next interactive git use), `git
    gc`/`repack`/`maintenance` as probe cases, and whether the hook's git read-ALLOWLIST
    (§4.2) rejects them; (2) which carve-out MEMBERS still auto-run beyond the enumerated 14
    (`sort`/`sed`/`file`/`readlink`/`strings`/`less`) — any that do get their own belt deny;
    (3) ADDED round 5 (R5-5): ATTACHED short-form flags against the HOOK allowlist — `git
    format-patch -o<path>` (no space; leaf-reproduced escaping every belt pattern) must be
    REJECTED by the hook's exact-read-form matching, leaf-tested not assumed — the hook is
    the close of record precisely because the belt is declared KNOWN-incomplete on this
    class, so its attached-form handling cannot remain an assumption.
    The bare `Bash` deny (R4-3) makes all of this belt-verification for rebuilt rungs, not
    the load-bearing close for the sleeper session. The classifier is a platform black box —
    test, don't trust.
19. Rung-2 Desktop tasks: can per-task permission config carry the full §4.2 profile and
    a wrapper equivalent? (The §3.4 gate-survival table marks these PARTIAL/UNKNOWN —
    R1-27.)
20. Can permission RULES scope the Workflow tool's `scriptPath` argument natively
    (`Workflow(scriptPath:...)` or equivalent), or is the sleeper-guard hook the only
    scoping mechanism for the debate.js channel? (R2-3; the hook path is the design of
    record either way — a native rule would be belt.)
21. Do stream-json deny events carry distinguishable PROVENANCE (PreToolUse hook-block
    message vs permission-rule denial)? (R2-2's belt witness; the fired-record is the
    load-bearing mechanism regardless — verify at build.)
22. Mid-drive structured output (R3-1): is `--json-schema` `structured_output` emitted
    for a MID-drive phase result under `--input-format stream-json`, or only for the
    final result (the only documented case)? Verify at build; the fenced-block
    text-parse fallback in §2.2 step 3 is the design of record until probed.
23. Phase-4 acceptance additions (R3-1 — behavioral legs of the two-phase drive, none
    yet probed): (a) second-message injectability — the drive can send the phase-1
    message after parsing phase-0 events; (b) phase-0 deny visibility — the canary
    denial appears as a parseable event in the output stream; (c) canary-attempt
    reliability — the model reliably ATTEMPTS the canary write when instructed (measure
    over N smoke runs; an unattempted canary aborts as no-fired-record, so the failure
    mode is availability, not safety). ADDED round 4 (R4-11): (d) the isolating
    dontAsk-zero-allow probe for the git `--output` gadget's approving layer — deferred to
    build because lens seats' nested `claude -p --permission-mode dontAsk` was denied by the
    seat's own auto-mode classifier; the bare `Bash` deny makes the outcome
    non-load-bearing, so this is attribution hygiene, not a gate dependency. ADDED round 5
    (R5-1): (e) the SEAT-SETTINGS-INHERITANCE probe — do debate.js-spawned Workflow seat
    agents inherit the headless `--settings` layer-1 profile (verify: a seat-attempted
    Bash call is absent/denied at the tool boundary in a headless smoke run)? The
    interactive hook evidence (layer 2) does not carry this; the design's seat-surface
    closure is inheritance-independent via the hook's Bash-channel deny (§4.3 layer 4
    (iii)), but the probe decides whether layer 1 ALSO binds the seats or the hook is the
    sole seat-surface layer — which changes the §6 row 13 defense-in-depth count, not the
    close.
24. RESOLVED round 5 by promotion (R5-3): quarantine is BUILT, in generalized
    positive-pool form — the corroboration pool is positively defined (§1.5 root
    invariant: pre-epoch date or interactive-origin marker; anything unattributable is
    quarantined counts-for-nothing and surfaced `unattributed`). The round-4 deferral had
    priced only the paste-run likelihood; red's rung-2 automatic-local-poisoning case
    (Desktop-task dirs counting as non-sleeper corroboration next morning with no human
    act) is what the deferral never priced and what forced the promotion. The
    loop-shaped-dir schema heuristic this question originally proposed is NOT built —
    provenance-keyed quarantine pays no schema false-positives.

---

## Footnotes (merged namespace; lane provenance noted per label)

[^HeadlessDocs]: "Run Claude Code programmatically" — Claude Code Docs,
  https://code.claude.com/docs/en/headless — accessed 2026-07-17 (full-page fetches, lanes
  1–3). Living source. Key quotes: "Add `--bare` to reduce startup time by skipping
  auto-discovery of hooks, skills, plugins, MCP servers, auto memory, and CLAUDE.md";
  "User-invoked skills and custom commands work in `-p` mode"; "`--bare` is the recommended
  mode for scripted and SDK calls, and will become the default for `-p` in a future
  release"; background workflows wait "capped at ten minutes by default"
  (`CLAUDE_CODE_PRINT_BG_WAIT_CEILING_MS`); `system/init` reports `plugins` and
  `plugin_errors`; `--output-format json` includes `total_cost_usd`; `--resume`/session
  ids; bare skips OAuth/keychain (API-key auth); 10MB stdin cap.
[^CliReference]: "CLI reference" — Claude Code Docs,
  https://code.claude.com/docs/en/cli-reference — accessed 2026-07-17 (lanes 1–3; lane 3
  cross-checked against `claude --help` on 2.1.212). Living source. Flags quoted verbatim:
  `--max-turns` ("Limit the number of agentic turns (print mode only). Exits with an error
  when the limit is reached"), `--max-budget-usd` ("Maximum dollar amount to spend on API
  calls before stopping (print mode only)"), `--disallowedTools`, `--permission-mode`
  (incl. `dontAsk`), `--mcp-config`, `--strict-mcp-config`, `--plugin-dir`,
  `--fallback-model` (round-1 correction R1-10: NO print-only marker on the current page,
  and a persistent `fallbackModel` setting is also documented), `--settings`,
  `--no-session-persistence`. Exit codes (round-1 correction R1-6): the page publishes no
  exit-code table; 0-on-success is probe-corroborated (P1/P2); the wrapper treats any
  nonzero exit as failure. Round-2 addition (R2-1): `--input-format <format>` — "text"
  (default) or "stream-json" (realtime streaming input), print-mode only — verified in
  `claude --help` on 2.1.212 (2026-07-17); the flag pair the wrapper's two-phase canary
  drive rides on. `--json-schema` invalid-schema exits with error ≥2.1.205 (red
  leaf-confirmed round 2).
[^PermissionsDoc]: "Configure permissions" — Claude Code Docs,
  https://code.claude.com/docs/en/permissions — accessed 2026-07-17 (full-page fetches,
  lanes 2–3). Living source. Carries: deny→ask→allow evaluation order with cross-level deny
  supremacy ("If a tool is denied at any level, no other level can allow it"); "Permission
  rules are enforced by Claude Code, not by the model"; `dontAsk` auto-deny semantics;
  Edit-not-Write path-rule matching (2.1.210 warning); Read-deny-blocks-Edit (v2.1.208);
  subprocess non-coverage warning + sandbox complementarity; blocking-hook precedence over
  allow rules; `-p` workspace-trust behavior ("no dialog appears and the rules stay
  ignored"); Windows path normalization (`C:\Users\alice` → `/c/Users/alice`);
  `disableBypassPermissionsMode` settable at any scope; settings precedence incl. managed
  settings above command-line arguments. Round-3 additions, re-fetched live 2026-07-17 at
  blue's round-3 verification (R3-14): the read-only Bash carve-out — "Claude Code
  recognizes a built-in set of Bash commands as read-only and runs them without a
  permission prompt in every mode. These include `ls`, `cat`, `echo`, `pwd`, `head`,
  `tail`, `grep`, `find`, `wc`, `which`, `diff`, `stat`, `du`, `cd`, and read-only forms
  of `git`. The set is not configurable; to require a prompt for one of these commands,
  add an `ask` or `deny` rule for it"; strict deny→ask→allow order with "a deny rule
  can't carry allowlist exceptions"; unquoted globs auto-run only for all-read-only-flag
  commands (`find`/`git` still prompt on unquoted globs); exec wrappers and `find
  -exec`/`-delete` always prompt.
[^HooksDoc]: "Hooks reference" — Claude Code Docs, https://code.claude.com/docs/en/hooks —
  accessed 2026-07-17 (lane 2). Living source. PreToolUse exit-2/JSON deny semantics;
  blocking-hook precedence (per permissions cross-reference); hook config file-watcher
  pickup; subagent `agent_id`/`agent_type` fields; plugin hooks.json source.
[^ScheduledTasks]: "Run prompts on a schedule" — Claude Code Docs,
  https://code.claude.com/docs/en/scheduled-tasks — accessed 2026-07-17 (lanes 1, 3).
  Living source. Cloud/Desktop//loop comparison table: /loop session-scoped, 7-day expiry,
  "No catch-up for missed fires"; Desktop: machine on, no open session, local files,
  per-task permission config; Cloud: no machine, no permission prompts, fresh clone,
  connectors, ≥1h interval. Also: "a scheduled fire only runs skills that Claude is allowed
  to invoke on its own... skills marked `disable-model-invocation: true`... reach Claude as
  plain text instead of executing."
[^RoutinesDocs]: "Automate work with routines" — Claude Code Docs,
  https://code.claude.com/docs/en/routines — accessed 2026-07-17 (lanes 1, 3). Research
  preview; explicitly volatile ("Behavior, limits, and the API surface may change"). Key
  quotes: "Routines run autonomously as full Claude Code cloud sessions: there is no
  permission-mode picker and no approval prompts during a run"; "By default, Claude can
  only push to branches prefixed with `claude/`"; each repo "cloned at the start of a run,
  starting from the default branch"; daily routine-run cap + subscription usage draw-down
  ("rejected until the window resets" absent credits); connectors all-included by default;
  actions attributed to the operator's linked identities; `/schedule` "requires a claude.ai
  subscription login. If ANTHROPIC_API_KEY... is set... remove it first"; "A green status
  in the run list... does not mean the task in your prompt succeeded."
[^HeadlessProbe]: Live headless probes P1/P2, this box, 2026-07-17, claude CLI 2.1.212
  (lane 3). P1: `claude -p "Reply with exactly: OK" --model haiku --output-format json
  --max-budget-usd 0.10` → exit 0, `total_cost_usd: 0.0246903`, `is_error: false`,
  `permission_denials: []`, `terminal_reason: "completed"`. P2:
  `claude -p "/prosthetic-conscience:probe" --model haiku --output-format json
  --max-budget-usd 0.30` → exit 0, plugin command executed, subagent spawned,
  critical-stance preload quoted verbatim, hook-test write permission-denied and reported;
  $0.058. Outputs quoted in the lane transcript; commands stated for re-derivation
  (ephemeral-instrument residue acknowledged — cheap fix: re-run and commit outputs under
  the run dir).
[^McpHeadlessBugs]: GitHub anthropics/claude-code issues (lane 3), status-checked
  2026-07-17: #76239 OPEN (stdio MCP tools silently missing on first turn when server start
  exceeds ~2s non-blocking pre-wait; regression since 2.1.144),
  https://github.com/anthropics/claude-code/issues/76239; #68375 OPEN (stdio tool call
  hangs with multiple servers loaded; `--strict-mcp-config` works around),
  https://github.com/anthropics/claude-code/issues/68375 — round-5 note (red live fetch
  2026-07-17): still OPEN with content zero-drift, but now carries a GitHub `stale` label
  alongside regression/has-repro, so bot auto-close is a live drift risk; a future CLOSED
  status would be bot lifecycle, not a fix claim — re-check at build; #32191 (`-p` with HTTP MCP server
  exits silently; 2.1.58–2.1.71 era), https://github.com/anthropics/claude-code/issues/32191
  — **CLOSED as duplicate** (canonical untraced), leaf-confirmed by red round 1 and
  adopted here, correcting round 0's search-listing carry (R1-5). Open ≠ will-be-fixed for
  the two open issues: design owns the workaround; #32191's phenomenon class stays on the
  checklist as history, not as an open bug.
[^PermAskBypass]: "[BUG] Edit/Write tools bypass permissions.ask rules (regression of
  #11226)" — anthropics/claude-code issue #22055,
  https://github.com/anthropics/claude-code/issues/22055 — accessed 2026-07-17 (direct
  fetch, lane 1). Status: **Closed as not planned**. Reproduction: Bash ask rules prompt;
  Edit/Write ask rules do not (files modified with no prompt), defeating protection of
  `.claude/hooks/**` and `.claude/settings.json`. Documented community workaround in the
  thread (verbatim, per red's gh --comments full-thread check): a PreToolUse exit-2
  protected-files hook. Round-1 correction (R1-13): chmod-444 is NOT in the thread —
  "chmod" appears once inside a commenter's allow-list snippet; the chmod-readonly measure
  in §4.3 layer 3 is this design's own proposal and is labeled as such there.
[^DenyRWIssue]: "Permission Deny Configuration Not Enforced for Read/Write Tools" —
  anthropics/claude-code issue #6631,
  https://github.com/anthropics/claude-code/issues/6631 — accessed 2026-07-17 (direct
  fetch, lane 1). Status: Closed; a prior fix was claimed via #4467, reporter re-confirmed
  the bypass at v1.0.93 (Aug 2025). Behavior in current builds unverified — treated as
  "cannot be the load-bearing layer" evidence, not a claim about today's build.
[^DenyBashIssue]: "Permission deny rules not enforced for Bash commands" —
  anthropics/claude-code issue #25621,
  https://github.com/anthropics/claude-code/issues/25621 — accessed 2026-07-17 (direct
  fetch, lane 1). Status: Closed as duplicate (phenomenon corroborated; canonical issue not
  traced — labeled accordingly).
[^WindowsHang]: "[DOCS] `claude -p` slow / appearing to hang on Windows during the
  slash-command and skill scan... (regression v2.1.161–v2.1.168, fixed in v2.1.169)" —
  anthropics/claude-code issue #66395,
  https://github.com/anthropics/claude-code/issues/66395 — accessed 2026-07-17 via search
  result title (lane 1; round 0 body-unfetched, MEDIUM). Upgraded MEDIUM → HIGH round 2:
  red fetched the issue body, which quotes the v2.1.169 changelog and the exact dated
  regression span v2.1.161–v2.1.168. Status leaf-confirmed round 1: CLOSED NOT PLANNED
  ([DOCS] class).
[^SlashHeadlessIssues]: anthropics/claude-code issues #837 ("use slash commands in
  print/headless/non-interactive mode") and #14246 ("Custom slash commands not discovered
  in CLI/SSH headless mode", v2.0.71, Linux/aarch64) — surveyed via search 2026-07-17
  (lanes 1, 2; issue open/closed status NOT individually fetched at round 0 — flagged per
  red Pattern A; statuses leaf-confirmed by red round 1: #837 CLOSED COMPLETED, #14246
  CLOSED DUPLICATE — the supersession story holds). Historical failure record superseded
  by the current headless doc's explicit support statement; retained as the reason the
  live acceptance test stays.
[^WebSandbox]: Claude Code on the web / sandboxing —
  https://code.claude.com/docs/en/sandbox-environments, Anthropic engineering post "Making
  Claude Code more secure and autonomous with sandboxing", and anthropics/claude-code issue
  #23707 ("Background Task agents fail on Claude Code Web — sandbox recycled between
  turns"). Surveyed 2026-07-17 (lane 2; disconfirming search against the cloud-default
  option). #23707 status leaf-confirmed by red round 1: CLOSED NOT PLANNED — the platform
  will not fix it, strengthening the not-the-default call.
[^GhaSchedule]: GitHub Actions `schedule` event behavior — GitHub community discussions
  #52477 and #156282 plus GitHub's documented "During periods of high load ... workflow
  runs may be delayed" / "may be dropped" guidance; 5-minute minimum interval; no SLA.
  Surveyed 2026-07-17 (lane 2); the delay/drop language is GitHub's own documentation, the
  numbers (5–30 min typical at :00) are community-measured.
[^MissedRun]: Missed-run tolerance primitives (lane 2): systemd.timer `Persistent=`
  ("saved to disk when they have been last triggered ... execute overdue timer events");
  anacron interval-since-last-run model; Windows Task Scheduler "Run task as soon as
  possible after a scheduled start is missed" (Microsoft Task Scheduler docs — the
  round-1 learn.microsoft.com task-scheduler SETTINGS page URL now 404s, red round-5
  note; re-pointed at the TaskSettings.StartWhenAvailable API reference page, which
  carries the same semantics and is the durable primary). Surveyed 2026-07-17.
[^UsageAPI]: "Usage and Cost API" — Claude Platform Docs,
  https://platform.claude.com/docs/en/manage-claude/usage-cost-api — accessed 2026-07-17
  via search digest (lane 1). `/v1/organizations/usage_report/messages` +
  `/v1/organizations/cost_report`; Admin API key required (org accounts); ~5-min freshness,
  1/min sustained polling. Upgraded MEDIUM → HIGH round 1 on red's leaf fetch (endpoints,
  Admin-key-only, "unavailable for individual accounts", freshness and polling figures all
  confirmed); the design claim it supports ("no subscription-auth monthly API guard") is
  additionally supported by the routines doc's auth constraints.
[^ConsoleLimits]: Anthropic Console workspace limits —
  platform.claude.com/docs/en/manage-claude/workspaces (workspace spend/rate limits
  settable below org limits) and anthropics/claude-quickstarts issue #371. Surveyed
  2026-07-17 (lane 2). Round-1 requalification (R1-9): the #371-derived "no endpoint to
  read or set" claim was stale on the READ half for RATE limits — see [^RateLimitsAPI];
  it stands for SPEND limits (no API read or set) and for anything reachable without an
  Admin key.
[^RateLimitsAPI]: "Rate Limits API" — Claude Platform Docs,
  https://platform.claude.com/docs/en/manage-claude/rate-limits-api — leaf-fetched by red
  lens 3, 2026-07-17; adopted round 1 (R1-9). `/v1/organizations/rate_limits` +
  `/v1/organizations/workspaces/{workspace_id}/rate_limits`: READ configured org and
  workspace rate limits; read-only; Admin API key required — unavailable to
  subscription-auth sessions, so the scheduler-cannot-poll design conclusion survives on
  auth grounds.
[^Pricing]: Model pricing and Batch API — canonical:
  https://platform.claude.com/docs/en/about-claude/pricing — leaf-fetched at red's round-1
  audit (2026-07-17), upgrading round 0's aggregator-carried MEDIUM (R1-11). Leaf figures:
  Haiku 4.5 $1/$5; Sonnet 4.5/4.6 $3/$15 (Sonnet 5 intro $2/$10 → $3/$15 from 2026-09-01);
  Opus 4.5–4.8 $5/$25; Fable/Mythos 5 $10/$50; Batch flat 50% off; cache reads 0.1×;
  new-generation tokenizer (Opus 4.7 and later Opus models, Fable 5, Mythos 5, Mythos
  Preview, Sonnet 5 — set completed round 3 per the page's own enumeration, R3-17: the
  round-2 list omitted the Opus 4.7+ members; completeness, not falsity — the named
  models DO use it) counts ~+30% more tokens than legacy counting. The ≤24h
  Batch async-window sub-claim is NOT on the pricing page; resolved HIGH round 2 at the
  platform batch-processing page ("Batches expire if processing does not complete within
  24 hours" — red live fetch 2026-07-17). Re-fetched live at red's round-2 audit with
  zero drift, same access date. VOLATILE — re-fetch at citation-verification.
[^SelfCorrect]: "Large Language Models Cannot Self-Correct Reasoning Yet" — Huang, Chen,
  Mishra, Zheng, Yu, Song, Zhou; ICLR 2024; arXiv:2310.01798,
  https://arxiv.org/abs/2310.01798 — accessed 2026-07-17 (lanes 1, 2). Intrinsic
  self-correction (no external feedback) fails to improve and sometimes degrades reasoning;
  prior claimed gains depended on oracle feedback. (Disconfirming search against
  introspective loop designs.)
[^Reflexion]: "Reflexion: Language Agents with Verbal Reinforcement Learning" — Shinn et
  al.; NeurIPS 2023; arXiv:2303.11366, https://arxiv.org/abs/2303.11366 — accessed
  2026-07-17 (lanes 1, 2, 3). Environment/execution feedback converted to persistent verbal
  memory (episodic buffer) consumed in later episodes; the working precedent for
  artifact-fed improvement loops.
[^Voyager]: "Voyager: An Open-Ended Embodied Agent with Large Language Models" — Wang et
  al.; arXiv:2305.16291, https://arxiv.org/abs/2305.16291 — accessed 2026-07-17 (lanes 2,
  3). Automatic curriculum + ever-growing skill library of executable, compositional skills
  built through "environment feedback, execution errors, and self-verification";
  "alleviates catastrophic forgetting."
[^DGM]: "Darwin Gödel Machine: Open-Ended Evolution of Self-Improving Agents" — Zhang, Hu,
  Lu, Lange, Clune; arXiv:2505.22954,
  https://arxiv.org/abs/2505.22954 — accessed 2026-07-17 (lane 2).
  Empirical-validation-over-proof framing; archive of agents. Round-1 corrections (R1-3):
  the "improve themselves the more compute they are provided" quote is NOT on the arXiv
  abs or /html pages — it is verbatim at sakana.ai/dgm/ and now lives in [^DGMSakana];
  the round-0 "(ICLR 2026)" venue tag is not stated on the cited page and is dropped.
[^DGMSakana]: "The Darwin Gödel Machine: AI that improves itself by rewriting its own
  code" — Sakana AI (authors' own project post — primary for the safety incidents),
  https://sakana.ai/dgm/ — accessed 2026-07-17 (lane 2). Quotes: sandboxed evaluation
  "under human supervision"; fake test logs; "it removed the markers we use in the reward
  function to detect hallucination (despite our explicit instruction not to do so)";
  detection via "transparent, traceable lineage of every change." Also verbatim here
  (moved from [^DGM], round-1 correction R1-3): agents "improve themselves the more
  compute they are provided."
[^SICA]: "A Self-Improving Coding Agent" — Robeyns et al., University of Bristol;
  arXiv:2504.15228, https://arxiv.org/abs/2504.15228 —
  accessed 2026-07-17 (lane 2). Venue cited as the page states it (round-1 correction
  R1-4): the abs page's Comments field says "Submitted as a preprint to NeurIPS 2025" —
  the round-0 "(ICLR 2025 SSI-FM workshop)" tag was not carried by the cited page and is
  dropped. Single agent as improver+improvee; archive-driven proposal
  step; 17–53% gains on a SWE-bench Verified subset.
[^STOP]: "Self-Taught Optimizer (STOP): Recursively Self-Improving Code Generation" —
  Zelikman et al.; arXiv:2310.02304 (COLM 2024), https://arxiv.org/abs/2310.02304 —
  accessed 2026-07-17 (lane 2). Seed improver improving itself; §Circumvention Attempts:
  GPT-4 unsandboxing 0.42% (CI 0.31–0.57%), 0.46% (CI 0.35–0.61%) with added warning —
  insignificantly HIGHER, two-proportion z-test; 10,000 sampled improvements, syntactic
  detection. Round-1 upgrade: figures re-pinned at ar5iv §6.2/Table 2 by three independent
  red lenses — OQ8 resolved at LaTeX-render fidelity; PDF pin optional publisher-grade
  polish.
[^IdeaStudy]: "Can LLMs Generate Novel Research Ideas? A Large-Scale Human Study with 100+
  NLP Researchers" — Si, Yang, Hashimoto; arXiv:2409.04109,
  https://arxiv.org/abs/2409.04109 — accessed 2026-07-17 (lane 1). LLM ideas judged more
  novel (p<0.05) but weaker on feasibility; weak diversity and unreliable self-assessment;
  human re-ranking helps — the evidence base for human triage at the idea gate.
[^Dependabot]: "Automating Dependency Updates in Practice: An Exploratory Study on GitHub
  Dependabot" — He et al.; arXiv:2206.07230, https://arxiv.org/abs/2206.07230 — accessed
  2026-07-17 (lane 2). Developers configure toward fewer notifications; 11.3% of projects
  deprecated Dependabot. (Disconfirming search against uncapped daily proposal cadence.)
[^DependabotFatigue]: "Reducing Alert Fatigue via AI-Assisted Negotiation: A Case for
  Dependabot" — arXiv:2502.06175, https://arxiv.org/abs/2502.06175 — accessed 2026-07-17
  (lane 2). Frames automated dependency PRs as an alert-fatigue problem (>75M PRs generated
  in 2022).
[^AIScientist]: "The AI Scientist: Towards Fully Automated Open-Ended Scientific
  Discovery" — Lu et al., arXiv:2408.06292 + Sakana AI announcement
  https://sakana.ai/ai-scientist/ — accessed 2026-07-17 (lanes 1, 3; round 0 at
  search-digest + abstract/post level). The system edited its own execution script to
  extend its timeout and relaunched itself recursively; authors recommend
  sandboxing/isolation. Realized precedent for an autonomous loop modifying its own
  constraints. Upgraded MEDIUM → HIGH round 1: red leaf-fetched sakana.ai/ai-scientist
  and confirmed the incident quotes verbatim ("edited the code to perform a system call to
  run itself... endlessly calling itself"; "modify its own code to extend the timeout
  period"; "mitigated by sandboxing").
[^AIControl]: "AI Control: Improving Safety Despite Intentional Subversion" — Greenblatt,
  Shlegeris, Sachan, Roger; arXiv:2312.06942 (ICML 2024),
  https://arxiv.org/abs/2312.06942 — accessed 2026-07-17 (lane 2). Safety protocols
  evaluated under intentional subversion by the untrusted model.
[^OWASP]: "Excessive Agency" — OWASP Top 10 for LLM Applications (LLM06:2025 lineage) —
  accessed 2026-07-17 (lanes 1, 2; lane 1 via secondary explainers — MEDIUM confidence on
  exact taxonomy wording, HIGH on the doctrine; lane 2 via genai.owasp.org mirror
  coverage). Root causes: excessive functionality/permissions/autonomy; mitigations: least
  privilege, human-in-the-loop approval for high-impact/sensitive actions, authorization in
  external systems, logging and rate-limiting tool invocations, draft-vs-execute
  separation.
[^Goodhart]: Reward hacking / specification gaming — "Reward hacking," Wikipedia
  (CoastRunners case, orig. OpenAI "Faulty reward functions in the wild," 2016),
  https://en.wikipedia.org/wiki/Reward_hacking; survey "Reward Hacking in the Era of Large
  Models," arXiv:2604.13602 — accessed 2026-07-17 at search-digest level (lane 1);
  qualitative use only, no figures carried.
[^AlertFatigue]: Alert-fatigue/actionability claims — practitioner/vendor literature
  (openobserve.ai, site24x7, ennetix AIOps posts) — accessed 2026-07-17 at search-digest
  level only (lane 1); the specific "under 1 in 5 alerts acted on" figure is NOT
  leaf-verified and is carried as qualitative direction only. LOW confidence on numbers;
  MEDIUM on the qualitative phenomenon (independently attested across sources).
[^CostRecord]: `research/2026-07-14_efficiency-investigation/cost.md` at pin `7bc501e`
  (lanes 1, 2, 3). Measured: total $414.97 list-rate across 42 agents / 1975 api-turns /
  4 rounds; cache traffic 99% of tokens; judgment-seat premium cache-RATE-driven; per-round
  red-lens lines $41.89–$61.46. Run-3 baseline $149.95 per `plans/efficiency-phase.md` §I
  (lane 3).
[^FrictionRun3]: `research/2026-07-12_feov-retrospective/friction.md` at pin `7bc501e`
  (lanes 1, 2, 3). 17 attributed entries; PDF extraction reported by red-merge every round
  (entries 1, 5, 7, 11, 17-adjacent); write-block filename-keyed isolation (entry 4);
  write-guard entries 3/8/10; Grep count footgun (12); Read-cap (15).
[^FrictionRun4]: `research/2026-07-14_efficiency-investigation/friction.md` at pin
  `7bc501e` (lanes 1, 2, 3). ~30–39 entries incl.: red gap-pattern memory unreadable at
  four seat classes; Read cap at six seat classes / every full-read seat every round;
  write-guard at five consecutive round-seats; MUST-try clause skipped live
  (blue-respond-r1); log() settled console-ephemeral only by ~/.claude spelunking
  (blue-respond-r2); ABORT DISCLOSURE (monthly spend-limit death, resumable state, "NO
  REPORT ASSEMBLED — resumable via wf_5cefd2a4-35f").
[^Backlog]: `ideas/backlog.md` at pin `7bc501e` (lanes 1, 2, 3). Items cited: 10
  (trajectories evaporate — "primary self-learning input"), 15–18 (smoke mode ~50k tokens;
  assemble-on-failure — range corrected round 2, R2-20: line 18 carries the
  assemble-on-failure sub-claim; content was verified verbatim, the round-1 range stopped
  one line short), 27c (PDF gap "requested by red, blue, AND judge across all 4
  rounds"), 28 (cost findings; panel counter excludes cache = 92% of flow), 29
  (lineage-blind detector → supersedes fix), 34 (qmd measured ladder), 36 (empty-bin hook
  crash-storm + bootstrap guard), 39 (batching prose ignored at haiku, 0/175).
[^QmdDaemon]: `ideas/backlog.md` item 34 at pin `7bc501e` (lanes 1, 2, 3): "HTTP DAEMON
  VERIFIED LIVE (2026-07-14, this box, CPU-only): `qmd mcp --http --daemon` works as README
  documents (PID file, `qmd mcp stop`, `/health`, MCP Streamable HTTP at :8181/mcp) —
  Phase 4 can depend on it." Measured ladder: bare-CLI hybrid 36.3s (model loading); daemon
  lex 2.9s; BM25 CLI 0.6s; MCP `query` takes client-authored searches arrays.
[^QmdFallback]: `research/2026-07-14_efficiency-investigation/friction.md` at pin
  `7bc501e`, blue-lane-1 entry (lane 3): qmd MCP unavailable at seat → "fell back to
  Grep/Read on the local corpus, workable here."
[^IdeasCorpus]: `ideas/backlog.md` and `ideas/doubts.md` at pin `7bc501e` (lanes 1, 3).
  Backlog: 25 statused checkbox items across 39 lines, with run provenance (round-1
  recount at the pin, R1-1 — the round-0 "40" reproduced nowhere); batching A/B "0/175
  multi-call messages" at haiku. Doubts: hypothesis → adjudication lifecycle, five founding doubts closed with
  evidence; closed item 3: "Plugin hooks + agent memory in workflow agents — CONFIRMED...
  sc-quality-gate fired on workflow-agent writes; red-auditor wrote its memory: project
  gap-pattern file."
[^EffReport]: `research/2026-07-14_efficiency-investigation/report.md` at pin `7bc501e`
  (lane 1), §1/§2.5/§6: severity-floor termination REJECTED as specified; per-round
  board-profile telemetry + documented stop-and-resume ratified; durable sink = merge-seat
  append to git-tracked `trajectories/board-telemetry.jsonl` with named consumers; "log()
  is console-ephemeral."
[^EfficiencyPlan]: `plans/efficiency-phase.md` at pin `7bc501e` (lane 3): ratified
  telemetry line (PR-A.1), red gap-pattern mirroring into run inputs (PR-C.2), attestation
  ceiling (§II constraints — in-run checks catch shape, post-hoc audit catches vacuity),
  bulk-tier freight note (§I out-of-scope), named revisit triggers.
[^ResearchCommand]: `plugins/frank-exchange-of-views/commands/research.md` at pin
  `7bc501e` (lane 3): `--smoke` parameters (1 lane, 1 round, haiku, ~50k tokens);
  keeper-run model guidance ("for keeper runs, omit `model` entirely"); stop-and-resume as
  standing practice; capture step emitting cost.md and run-record-audit.md; "Prose here is
  for DECISIONS; the mechanics are scripted (design-by-contract: an LLM executing mechanics
  is an unenforced good-faith contract)." Round-2 execution-locus probe (R2-3), read at
  the shipped plugin copy 0.7.0 (= the pinned command's shape), command-file steps 2/3/5:
  step 2 runs `node .../setup-research-run.mjs` and step 5 runs
  `node .../capture-research-run.mjs` (session-Bash invocations); step 3 "Invoke the
  **Workflow** tool with `scriptPath` = `.../debate.js`" (harness-side runner, not Bash) —
  the mixed locus §2.2 step 4 and §4.3 layer 4 now state.
[^SmokeRecord]: `research/2026-07-17_smoke-ab-memarch-review/friction.md` at pin `7bc501e`
  (lane 3): single-round UNVERIFIED assembly with Catechism template misfit surfaced as
  friction — the bounded mode's honest-artifact precedent.
[^SemanticConsent]: `plugins/prosthetic-conscience/skills/semantic-consent/SKILL.md` at
  pin `7bc501e` (lane 3), final clause quoted verbatim in §4.3.
[^PushGuard]: `plugins/prosthetic-conscience/tools/cmd/sc-push-freeze-guard/main.go` at
  pin `7bc501e` (lane 3), contract comment: "It NEVER blocks — the freeze is a commitment
  the human may consciously override; the guard's job is making the commitment impossible
  to forget, not impossible to break."
[^HooksJson]: `plugins/prosthetic-conscience/hooks/hooks.json` at pin `7bc501e` (lane 3):
  every hook command wrapped in the bootstrap guard ("a fresh plugin-cache version ships
  from git WITHOUT binaries … an unguarded hook crash-storms every tool call in that
  window").
[^RedPatterns]: `inputs/red-gap-patterns.md` (this run's staged mirror of red's
  gap-pattern memory, 1,557 lines — byte-exact recount round 1, R1-30: final byte is 0x0a,
  terminated last line; lanes 1, 2, 3 read pre-flight):
  invariant-soundness-by-enumeration (denylists under-include; recommend allowlist
  inversion) applied to §4's gate shape; citation Pattern A (issue-status checks), Pattern
  B/E (figure-to-source fidelity), live-source drift, gitignored≠absent, file-type
  blindspot, policy-without-mechanism, ephemeral-instrument — all applied in lane methods.
[^PortPlan]: "Claude Code port plan" §3c + Phase table — read at
  `AgentOrange/docs/claude-port-plan.md` (lane 3 at `6df52af`; lanes 1–2 from the working
  tree) because the pin's `plans/claude-port-plan.md` path does not exist in the
  special-circumstances tree at `7bc501e` (verified by `git show`; standing friction —
  cross-corpus citation is snapshot-grade, not pin-grade). Red re-verified round 1 (R1-7):
  the path is confirmed absent at the pin, so the run's own PINNED.md asserts a
  nonexistent path — a run-infrastructure defect (setup-script pin validation / stage the
  port plan into inputs/), routed to the lead; the quotes themselves verified verbatim
  against the working tree (MEDIUM, snapshot-grade). §3c: sleeper-service structure
  (continuous-learning skill, self-improve.md, graduate.md, docs/scheduling.md); guardrail
  "the loop writes only research/ and ideas/; promotion into rules/skills requires the
  human (Semantic Consent)"; "human approves each step"; Phase-4 verify: "Headless
  `claude -p \"/self-improve\"` produces a run dir + idea stub; touches only
  research/+ideas/" (the plan's verify step is SUPERSEDED as-written round 5, R5-2 —
  under the R4-1 trampoline that fire is inert by design and produces no run dir; the
  adopted two-legged restatement lives in §3.3 item 1: wrapper `--manual` produces the
  run dir, the command produces NONE); resolved decision 6: daily cadence, scheduling
  always human-opt-in.

---

## Red team findings (in full)

*Red's board embedded verbatim: `red/ledger.md` (single source of truth for gap status — the final FAIL verdict and all 10 open gaps) followed by `red/archive.md` (the full closure record, every supersedes chain end-to-end). Per-lens evidence is preserved under `red/candidates/round-{1..5}-lens-{1..6}.md`; the ~leaf-verified statement↔reference record is `red/citation-ledger.md`.*

### red/ledger.md (verbatim)

---

# red/ledger.md — sleeper-service design run (SINGLE SOURCE OF TRUTH for gap status)

Round 5 merge, 2026-07-17. Lens passes: L1–L4 (leaf citation verification, 4 slices),
L5 (logic & completeness), L6 (dark-side & risk). Candidate files preserved under
`red/candidates/round-5-lens-*.md`; found_by is auditable against them.
Grading legend: likelihood × impact × complexity-to-mitigate → severity.
"likelihood: certain" on textual defects = the defect is verified present (the mass
mapping's convention); consequence-likelihood is carried in impact.

Round-5 headline: blue's round-4 revision addressed all 16 round-4 gaps (bare `Bash` deny,
git hook read-allowlist, /self-improve trampoline, memory-deny declaration, R4-9 root
invariant, plus the cheap fan). Round-5 lenses re-read the full report (2,159 lines, whole,
consecutive windows) and re-fetched the volatile leaves live (scheduled-tasks, routines,
permissions, hooks, pricing, missed-run, GHA, web-sandbox docs; the GitHub issue set;
IdeaStudy) — zero refuted external leaves; ten living-source claim sets zero-drift.
Disposition: **6 R4 gaps CLOSED clean** (R4-5/7/10/11/15/16), **10 CLOSED WITH REGRESSION**
(successors below, declared lineage). R1-7 stands ADJUDICATED (lead risk_accepted round 2;
excluded from the verdict). The round-5 board is 10 residues. The round's shape: round 4's
STRUCTURAL closes hold at the top-level session but were not derived over the two surfaces
where nightly work actually executes — the Workflow-spawned FEOV seat agents (R5-1) and the
negatively-defined corroboration pool (R5-3) — plus repair-propagation and spec-totality
residues, all cheap. Mass 30.0 (round 4: 46.0; round 3: 55.0), max severity medium, open
count 10 — converging.

## OPEN GAPS (10)

### R5-1 — R4-3's "the session never invokes Bash" and R4-2's "Where Bash IS reachable (… the Workflow seat agents …)" contradict on whether the nightly FEOV seat agents sit inside the bare-`Bash`-deny boundary — and BOTH horns are defective: bound ⇒ the nightly pass's seats silently lose Bash (a capability the design's own citation/live-probe seat methods demonstrably use) and R4-2's premise is false; not bound ⇒ the "closed at the tool boundary" class-closure claims are void for the actual nightly worker population, and seat inheritance of the headless `--settings` profile is asserted from interactive-run evidence only
- found_by: [L3, L5, L6]
- supersedes: [R4-3, R4-2]
- location: §4.2 R4-2 bullet — "Where Bash IS reachable (a rebuilt rung, the Workflow seat
  agents, profile drift), the round-3 git posture … was still a denylist"; §4.2 R4-3 bullet —
  "§2.2's session steps never invoke Bash — git and node run WRAPPER-side — so denying Bash
  costs the session nothing and closes … the whole class"; §4.3 layer 4 (iii) — "the
  workflow's SEAT AGENTS are full permission-engine + hook subjects"; §6 row 13 — "the Bash
  read carve-out is closed STRUCTURALLY round 4 by the bare `Bash` deny — R4-3 — so the
  R1-17 read-scoping now holds on the Bash channel too"; §4.2 git allow comment — "the
  built-in read-only-git carve-out auto-approves read-only git regardless (R3-14)".
- problem, three composing legs. (a) INTERNAL CONTRADICTION, certain-present: layer 4 (iii)
  makes seat agents full permission-engine subjects (⇒ the bare deny binds them ⇒ Bash NOT
  reachable), while the same round's R4-2 bullet names "the Workflow seat agents" as a
  surface where Bash IS reachable. Both cannot hold. (b) IF SEATS ARE BOUND: "denying Bash
  costs the session nothing" is argued only over §2.2's eight top-level steps, never over the
  seat population step 4 spawns nightly — this design's own evidentiary record is that
  citation/verification seats run Bash git probes routinely (the R3-15/R4-2/R5-5 gadget
  reproductions were exactly that), and lane-3's live-probe method is Bash-dependent; a seat
  needing Bash fails-denied every night, its friction is ordinary-classed and waits forever
  under R2-6 on non-sleeper corroboration — silent, unpriced capability starvation of the
  nightly pass, the shape R4-4 owned for memory writes and nowhere owns for Bash. (c) IF
  SEATS ARE NOT BOUND: §6 row 13's and §4.3 layer 2's "closed at the tool boundary … holds
  on the Bash channel too" TOTAL claims are overclaimed at the one surviving script channel;
  closure for seats reverts to the belt enumeration R4-3 itself declared non-load-bearing
  ("include[s]", non-exhaustive), and the fence covers WRITES, not Bash READS of box secrets
  — the R1-13 read+egress channel re-opens on the seat surface. Worse: the seat-coverage
  leaf evidence (sc-quality-gate fired on workflow-agent writes) is from interactive runs;
  whether debate.js-spawned seats inherit the headless session's `--settings` layer-1
  profile AT ALL (bare Bash deny, Read-scoping, WebSearch drop, arxiv-only egress) is
  unprobed — if seats run under default settings (`defaultMode: auto` on this box, the very
  layer-masking fact R4-11 established), NONE of §4.2 binds the nightly worker population.
  Hook coverage (layer 2) is proven for seats; settings inheritance (layer 1) is silently
  assumed. Sub-leg (trivial, L3 live-doc refutation): the four retained `Bash(git …)` allow
  rules are dead under deny supremacy, and the retained comment "auto-approves read-only
  git regardless" is REFUTED by the same footnote the profile rests on — the live doc says
  a bare `Bash` deny "removes the tool from Claude's context entirely," so the within-Bash
  carve-out is vacuous under the shipped profile; the comment is the un-reconciled
  R3-14-era survivor of the R4-3 edit.
- required_fix: state which horn holds and re-derive the dependent text. If seats are bound
  (safe polarity): correct R4-2's reachability list (drop "the Workflow seat agents" or
  re-scope to rebuilt rungs/drift); own the nightly seat-Bash capability cost in §2.2 step 4
  / §5.2 (smoke-lane and citation-pass methods Bash-free, or the loss priced); re-label the
  dead git allow rules and drop/rung-caveat the "auto-approves regardless" comment. If seats
  are not bound: re-scope §6 row 13, §4.2, §4.3 layers 2/4, and §6 row 4 leg (a) to the
  top-level session, and extend the fence to Bash READS on the seat surface or generalize
  the R4-2 hook read-allowlist to the whole Bash channel per invariant 6. EITHER WAY: add
  the seat-settings-inheritance probe as a named OQ23 acceptance leg — layer 4 (iii)'s
  interactive evidence does not carry the headless `--settings` case.
- grading: medium (contradiction verified present; the operational question live either
  way; weaponizing leg (c) needs injection reaching a seat — but that is row 13's already
  accepted threat) × medium-high (horn (c) voids §4.2 for the nightly workers; horn (b) is
  silent capability starvation + a false premise in a load-bearing bullet) × low-medium
  (one stated decision + one OQ leg + comment re-label; the generalized allowlist is a
  matcher change) → severity **medium**

### R5-2 — R4-1 moved the /self-improve payload out of `commands/`, but two body sites still specify the OLD shape: §3.4's rung-0 ladder cell ("the /self-improve command markdown is the wrapper's phase-1 prompt payload in EVERY mode") and §3.3's adopted Phase-4 acceptance test (`claude -p "/self-improve"` produces a run dir) — a test the trampoline design now FAILS BY CONSTRUCTION, whose cheapest pass is re-inlining the payload (undoing R4-1)
- found_by: [L2, L5]
- supersedes: [R4-1]
- location: §3.4 ladder row 0 — "the /self-improve command markdown is the wrapper's
  phase-1 prompt payload in EVERY mode, not a standalone entry point (R3-2)"; §0 tree —
  "the full loop PAYLOAD lives in the wrapper's phase-1 prompt (skills/continuous-learning),
  NOT in commands/"; §3.3 item 1 — "the port plan's Phase-4 verify step ('Headless `claude
  -p "/self-improve"` produces a run dir + idea stub; touches only research/+ideas/') must
  remain the acceptance test."
- problem: (a) the rung-0 cell asserts the round-3 shape the R4-1 edit list never covered —
  the ladder table is scheduling.md's shipped artifact, and a builder reading row 0
  re-creates a payload-carrying invocable command, the exact vector R4-1 closed
  (incomplete-repair body-lag; the round-4 propagation greps carry no token covering
  "phase-1 prompt payload"). (b) Under R4-1, `claude -p "/self-improve"` reaches Claude as
  the trampoline's instruction text — no run dir is produced, so the adopted acceptance
  test is unsatisfiable as written; a builder treating it as the Phase-4 gate either fails
  the gate or quietly re-inlines the payload to make it pass. The correct post-R4-1 test is
  TWO-legged: (i) `node sleeper-wrapper.mjs --manual` produces the run dir + stub touching
  only research/+ideas/; (ii) `claude -p "/self-improve"` produces NO run dir (trampoline
  inertness — the R4-1 property itself becomes verifiable).
- required_fix: rewrite the rung-0 cell to the R4-1 shape (thin trampoline; payload is the
  wrapper's phase-1 prompt sourced from the skill file); restate the Phase-4 acceptance
  test in §3.3 as the two-legged post-trampoline form (port-plan quote kept as historical
  source); add "phase-1 prompt payload" to the propagation-grep list.
- grading: certain (both lags verified as printed) × low-medium (the named Phase-4 gate is
  unsatisfiable as written; perverse incentive undoes a round-4 repair) × trivial (two
  clause edits) → severity **low-medium**

### R5-3 — the corroboration pool is defined NEGATIVELY (everything not tagged `origin: sleeper` corroborates) and round 4 patched three feeder surfaces in isolation (R4-1 paste-run, R4-6 unobserved-window dirs, R4-13 rung-2/3/4 markerless dirs) without the root fix R4-1's own required_fix named — the pool is now three rounds deep in per-surface patches (R2-6 → R3-3/R3-5 → R4-1/R4-6/R4-13) with no root invariant, and the sharpest surface (rung-2 local Desktop task) makes the poisoning AUTOMATIC behind an instructional-only adoption gate
- found_by: [L6]
- supersedes: [R4-13, R4-1, R4-6]
- location: §1.5 — "Sleeper-added content is capped at 1 per class and excluded from the
  non-sleeper corroboration pool"; §3.4 gate-survival R4-13 row — "NO (wrapper absent →
  sleeper dirs land markerless → next harvest counts them non-sleeper)"; §3.4 (R4-1
  polarity) — "a markerless dir is classed NON-sleeper, so its friction does not NEED
  corroboration — it PROVIDES the non-sleeper corroboration"; §8 OQ24 — "Deferred, not
  built."
- problem: pool membership is default-negative — a dir corroborates unless affirmatively
  tagged. Each round-4 patch is a per-surface bound (trampoline-makes-harder for R4-1;
  next-START window close for R4-6; a table row + adoption-requirement PROSE for R4-13);
  none adopts the positive-pool/quarantine root fix R4-1's required_fix (iii) named and
  OQ24 defers. This is the exact per-surface chain shape blue ended for status timers with
  R4-9's root invariant, declined here. Sharpest instance: rung 2 (Desktop scheduled task)
  runs LOCALLY against the SAME corpus the nightly harvest reads — its markerless sleeper
  dirs are counted NON-sleeper next morning AUTOMATICALLY, no human merge in the path, and
  the only guard is an instruction written into the adoption stub — the gate class §4.1's
  own 0/175 headline evidence refutes. OQ24's deferral rationale prices only the paste-run
  likelihood (a human ignoring their own contract), not the rung-2 automatic case.
- required_fix: state the root invariant once (mirroring R4-9): the corroboration pool is
  defined POSITIVELY — a dir corroborates only with affirmative non-sleeper provenance;
  anything harvest cannot positively attribute is quarantined (neither sleeper-capped nor
  corroboration-eligible), fail-closed toward counts-for-nothing — OQ24's quarantine
  promoted from deferred to built, dissolving three per-surface residuals in one predicate.
  If blue holds the deferral, the risk-accept must argue the rung-2
  automatic-local-poisoning case specifically.
- grading: low-medium (rung-2 adoption is human-gated, but once adopted the poisoning is
  automatic; the other two surfaces individually low) × medium (the ordinary-class
  self-poisoning bound voided on any surface feeding the pool) × low (one harvest
  predicate — cheaper than the three per-surface bounds it replaces) → severity **medium**

### R5-4 — R4-6's confinement clause re-introduces the dir-NAME keying R2-5 abolished, contradicting §1.5's still-standing "harvest.mjs reads ONLY the marker, never the dir name" — and the named convention "the wrapper's own sub-run slug" is not knowable: the slug is model-chosen per night, format-identical to human run dirs, and after the hard-kill the clause exists for, nothing durable records what that night's slug was
- found_by: [L5]
- supersedes: [R4-6]
- location: §1.5 — "harvest.mjs reads ONLY the marker, never the dir name"; §1.5 R4-6
  clause (ii) — "its markerless sweep is CONFINED to dirs bearing the sleeper date-key
  naming convention (`research/<date>_self-improve/` and the wrapper's own sub-run slug)";
  §6 row 10 — "the unobserved-exit window sweep is confined to sleeper date-key naming so a
  hard-kill cannot sweep human-present dirs (R4-6)."
- problem: (a) DOCTRINE vs MECHANISM: the unqualified round-2 no-name-reads sentence stands
  in the same section whose round-4 clause decides sweep membership BY name — inside a
  retroactive-uncertain window, date-key-named markerless dirs ARE auto-tagged sleeper by
  name. (b) `<date>_self-improve/` is a real static convention; "the wrapper's own sub-run
  slug" is not — sanitized from the model's phase-3 pick, different every night,
  pattern-indistinguishable from a human's same-day research dir. In the hard-kill scenario
  the wrapper that knew the slug is dead: the confinement either cannot match the sleeper
  sub-run (goes to human confirmation — acceptable, but then row 10's auto-tag claim
  under-delivers) or must match `<date>_*` (sweeps human dirs — the exact harm (ii) was
  minted to prevent). Mitigating fact the text never states: the marker is stamped AT
  CREATION, so a markerless sleeper sub-run exists only in the mkdir-to-stamp instant.
- required_fix: (i) qualify the §1.5 doctrine sentence (name-keying permitted only to
  CONFINE retroactive-uncertain sweeps, never to assign origin outside one; scope row 10
  accordingly); (ii) make the slug knowable — the wrapper appends the sub-run dir PATH to
  the run-window log AT CREATION, beside the START record; confinement then matches
  recorded paths and the doctrine sentence survives intact; (iii) state the mkdir-to-stamp
  bound explicitly.
- grading: low (hard-kill inside the stamp gap, or same-day human dir inside a retroactive
  window) × low-medium (corroboration-pool escape or the human-sweep harm; plus a standing
  doctrine contradiction) × low (one log line + two clauses) → severity **low-medium**

### R5-5 — R4-2's belt extension is STILL non-exhaustive on the exact class it was minted to close: the ATTACHED `-o<value>` form (no space) escapes `Bash(* -o *)` — merge-reproduced this box (`git format-patch -1 -o/tmp/r5mergeA HEAD` → exit 0, arbitrary out-of-repo patch) — the enumerate-and-extend regress recurring one lexical form deeper INSIDE the R4-2 repair; belt-only, but the belt is what binds rebuilt rungs 3–4
- found_by: [L6]
- supersedes: [R4-2]
- location: §4.2 deny block — `"Bash(* -o *)", "Bash(* -O *)"` (R4-2 comment: "`-o` matched
  none of the three long-form denies"); §6 row 4 leg (a) — "`git format-patch -1 -o <path>`
  → exit 0 arbitrary out-of-repo patch, R4-2 leaf-verified."
- problem: `Bash(* -o *)` requires space-delimited `-o`; git (like most getopt short
  options) accepts the attached form. Leaf-verified twice this round (L6 lens + merge
  re-run 2026-07-17): `git format-patch -1 -o/tmp/<dir> HEAD` → exit 0, patch written
  out-of-repo; no belt pattern matches. This does NOT reopen the session (bare `Bash` deny)
  and the hook git read-allowlist — the enforcement of record — rejects `format-patch`
  entirely IF its parser handles attached forms; but the belt binds rebuilt rungs 3–4, and
  the escape is direct evidence the enumerate-and-extend approach cannot close this class
  (the argument R4-3 itself made for the carve-out).
- required_fix: do NOT extend the belt pattern again; state in-file that the belt is
  KNOWN-incomplete on attached short forms and the hook allowlist is the actual close; add
  `git format-patch -o<attached>` to OQ18(c)'s probe matrix so the hook allowlist's
  handling of attached-form flags is leaf-tested, not assumed.
- grading: low (belt-only; hook allowlist + bare deny are the loads of record — but the
  rung-3/4 belt exposure is real) × medium (arbitrary out-of-repo write on any rung relying
  on the belt) × trivial (one in-file note + one OQ probe case) → severity **low**

### R5-6 — R4-4's `expected-memory-deny` bucket is keyed on the TARGET PATH, not the actor or cause: every `.claude/agent-memory/**` fence-denial — the nightly merge-seat's one expected protocol write AND an injection-driven memory-poisoning ATTEMPT (§6 row 14's Medium-likelihood channel) — lands in the one bucket the doctor line is told to ignore, at any volume; enforcement holds, but attack-attempt telemetry on the surface that steers red's future audits is erased by design
- found_by: [L5, L6]
- supersedes: [R4-4]
- location: §1.5 — "the resulting fence-denial is a NORMAL fired-record class the doctor
  line does not flag as anomaly … the wrapper's fired-record parser buckets
  `.claude/agent-memory/**` denials as `origin=nightly-seat, class=expected-memory-deny`."
- problem: the expected event has a known shape — the red-merge seat's record-new-patterns
  write, roughly once per run, targeting the red-auditor memory dir. A poisoned/injected
  session probing the memory surface produces the same PATH class at arbitrary volume and
  arbitrary targets within `agent-memory/**`, and all of it files as expected and
  unflagged. The design treats guardrail-surface tamper ATTEMPTS as must-be-visible (TAMPER
  persists until cleared; canary aborts surface); agent-memory — the corpus steering the
  adversary's future audits — gets its denials routed into the one suppressed bucket.
  Invariant 7's own test: a control that records without surfacing is telemetry. Red's
  round-4 question (are routine fence denials distinguishable from attack evidence?) was
  answered by DEFINING the class non-anomalous rather than by mechanism — the
  origin-tag-by-CAUSE lesson inverted onto a benign-bucket.
- required_fix: one predicate in the parser — `expected-memory-deny` ONLY for the nightly
  red-merge seat's own record-new-patterns write shape (seat role + pattern-file target),
  bounded ≤N (small) per run; any other actor, target subdir, or volume over the bound
  surfaces as a distinct `unexpected-memory-write-attempt` class on the doctor line
  (persists like TAMPER).
- grading: low-medium (needs an injection reaching an in-session actor; row 14 grades that
  channel Medium; the write stays blocked either way) × medium (attack visibility lost on
  the adversary-steering surface) × low (one actor/shape predicate + a count bound) →
  severity **low-medium**

### R5-7 — the SessionStart staleness hook now ships in sleeper-service's OWN hooks.json (R4-8) and fires in EVERY interactive session, but §6 row 9's crash-storm bound cites prosthetic-conscience's hooks.json — bootstrap-guard coverage for the NEW hooks.json is unstated, and R4-14's flag-ABSENT anomaly print nags every never-configured install (flag-absent is the DEFAULT state after install, before schedule-setup ever runs)
- found_by: [L6]
- supersedes: [R4-8, R4-14]
- location: §0 — "it must fire in interactive sessions on this box independent of the
  sleeper scheduler, so it ships with the sleeper plugin's own hooks.json, R4-8"; §6 row 9
  — "Bootstrap guard already shipped in hooks.json[^HooksJson]" ([^HooksJson] =
  prosthetic-conscience's hooks.json at the pin); §3.4 (R4-14) — "flag-absent prints
  `sleeper: flag missing; last check <ts>` (an anomaly)."
- problem: (a) row 9 bounds the empty-bin hook crash-storm with a guard proven for
  prosthetic-conscience; the sleeper SessionStart hook lives in a DIFFERENT hooks.json and
  fires in every interactive session, so an unguarded sleeper hooks.json crash-storms all
  interactive work during a sleeper-service cache update — a surface row 9 (nightly-scoped)
  never contemplated, and guard coverage for the new file is asserted nowhere. (b)
  flag-absent is the default state for any install that has not run schedule-setup; printing
  it as a per-session ANOMALY is the alert-fatigue mode the report's own Dependabot
  evidence catalogs, eroding the dead-man channel's readership — never-configured vs
  flag-lost are undistinguished.
- required_fix: (i) state that sleeper-service's hooks.json wraps its SessionStart command
  in the bootstrap guard and re-point row 9's bound to it (add the hook to the empty-bin
  acceptance check); (ii) gate the flag-missing anomaly on a prior `last-successful-run`
  record — no prior run ⇒ never configured ⇒ silent or one-time notice.
- grading: low (empty-bin window narrow; nag cosmetic-but-erosive) × low-medium
  (crash-storm reaches interactive work if unguarded; chronic false-anomaly erodes the
  channel R4-14 hardened) × trivial (one guard-coverage clause + one gate condition) →
  severity **low**

### R5-8 — §0's tree (labeled "the implementable shape") was never reconciled with the R4-8 count beside it: the FOURTH code artifact (SessionStart staleness hook) and the sleeper hooks.json that R4-8 homes it in appear nowhere in the printed tree — and nothing in the tree registers the sleeper-guard PreToolUse hook either
- found_by: [L5]
- supersedes: [R4-8]
- location: §0 tree (`.claude-plugin/`, `skills/`, `commands/`, `scripts/` (harvest.mjs,
  sleeper-wrapper.mjs, sleeper-guard), `docs/scheduling.md` — no `hooks/` or hooks.json
  entry) vs same section — "the **SessionStart staleness-warning hook** (a small executable
  + its hooks.json entry … it ships with the sleeper plugin's own hooks.json, R4-8)."
- problem: the section's two representations of the artifact set disagree; a builder
  implementing from the tree ships three of four counted code artifacts and no hook
  registration at all. The enumeration was repaired three times (R2-19, R3-10, R4-8) while
  the TREE never was — exhaustive-sweep-omits-own-specimen, one artifact deeper than R3-10.
- required_fix: add `hooks/hooks.json` (registering the sleeper-guard PreToolUse hook and
  the SessionStart staleness hook) and the SessionStart executable to the §0 tree.
- grading: certain (divergence verified) × low (prose complete; tree misleads a skimming
  builder) × trivial (two tree lines) → severity **low**

### R5-9 — R4-9's `rejected` re-surface trigger has no computable anchor: both arms of clause (b) — "default 90 days" and "until the class's recurrence rate exceeds its pre-rejection rate" — key on a rejection DATE no artifact records (stub filenames date the MINT; status edits are undated), and the `regression` token's domain is unpinned (docket flag vs status enum; the graduated stub's post-recurrence status and the `rejected-recurring` setter are unstated)
- found_by: [L2, L5]
- supersedes: [R4-9]
- location: §1.4 root-invariant clause (b) — "**`rejected`** — a rejected stub dedupes its
  class for a cadence-tuned window (default 90 days, operator-tunable, OR until the class's
  recurrence rate exceeds its pre-rejection rate, whichever first)"; §2.3 status enum —
  "open | stale | graduation-queued | queued-stale | graduated | rejected |
  rejected-recurring … humans set graduated/rejected" (no date-of-change field; `regression`
  not in the enum).
- problem: (a) computing either arm needs the rejection date to split the timeline; it
  exists nowhere — a builder implements filename-date+90d (measures mint age, not rejection
  age: a stub rejected at day 80 re-surfaces in 10 days) or silently drops the rate arm
  (printed invariant ≠ built invariant — the R4-12 defect class inside the R4-9 repair
  itself). (b) Of the four R4-9 re-surface flags, three are enum values but `regression` is
  not, and no rule states what the graduated stub's status becomes on recurrence or who
  sets `rejected-recurring` on stub vs docket — pinned-mapping-not-total on exactly the
  terminal states R4-9 was minted to govern.
- required_fix: one dated field (the status line records the last human status change,
  e.g. `status: rejected 2026-07-17`; both arms key on it; harvest parses it) + one clause
  pinning `regression`'s domain (e.g. docket-flag only; the stub stays `graduated`; the
  regression entry may mint a NEW stub) and the `rejected-recurring` setter.
- grading: certain (no anchor/source stated) × low (mis-timed re-surface windows; the
  invariant survives in intent) × trivial (one dated field + one clause) → severity **low**

### R5-10 — R4-12's "NAMED source" is an unparseable and currently-empty convention: "harvest parses that note's value" states no format (backlog entries are free prose), and leaf check shows NO pinned backlog entry carries any parseable complexity note — the divisor is universally inert against the actual corpus while the text implies it activates
- found_by: [L1, L5]
- supersedes: [R4-12]
- location: §1.4 stage 2 — "it defaults to **1 (inert — the factor vanishes)** unless the
  class's matching `ideas/backlog.md` entry carries a human-recorded complexity note, in
  which case harvest parses that note's value."
- problem: (a) no format stated — a zero-token parser over free prose either guesses (the
  defect R4-12 closed) or never matches (factor permanently inert); (b) L1 leaf check at
  the pin (`git show 7bc501e:ideas/backlog.md`, exhaustive grep for
  complexity/effort/difficulty tokens): no structured field on any of the 25 items — the
  named source is a corpus convention that does not yet exist. The safe default (1) bounds
  the damage; the text should say the field is forward-looking.
- required_fix: state the token convention in §1.4 stage 2 (e.g. a literal `cx:<1-5>` token
  in the backlog entry; anything else = default 1) and note the field is currently
  unpopulated (forward-looking curation surface).
- grading: certain (no format stated; source empty at the pin) × trivial (default-1 bounds
  it; the human sees the full ranked table) × trivial (one clause) → severity **trivial**
  (recommend, not block)

## CLOSURE INDEX

R1-1 | closed | backlog count corrected to 25/39, recount verified at pin (L1 HIGH) | —
R1-2 | closed | scope-fusion split into two attributed sources, both verified at pin lines 27c/31h (L1 HIGH) | —
R1-3 | closed | Sakana quote moved to [^DGMSakana]; ICLR-2026 tag dropped (L1 verified) | —
R1-4 | closed | [^SICA] venue now as the page states; re-fetched verbatim r2 (L1 HIGH) | —
R1-5 | closed | #32191 restated CLOSED-duplicate in §3.3/§6/footnote; propagation grep clean (L2, merge grep) | —
R1-6 | closed | exit-code claim softened to any-nonzero-is-failure; cli-reference re-fetched r2, still no exit table (L2) | —
R1-7 | risk_argued | LEAD risk_accepted round 2: valid+verified (git cat-file MISSING at 7bc501e) but fix is run-infra owned by lead; port-plan citations remain snapshot-grade, quotes re-verified verbatim at 6df52af | —
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
R1-21 | closed_with_regression | row 4 re-argued without actor-benignity; leg (a) premise contested by step-4 locus gap | R2-3
R1-22 | closed_with_regression | 30-day auto-stale mechanism specified in §1.4/§2.3; composes badly with graduation latency | R2-11
R1-23 | closed | Batch demoted to FUTURE note; ≤24h sub-claim since resolved HIGH on the batch-processing page (L3 V2) | —
R1-24 | closed | OQ3 qualifier carried inline in layer-6 table row (merge direct read) | —
R1-25 | closed_with_regression | §1.5 covers authorship, origin-tag + cap + corroboration gate specified; tag name-keyed and disjunction undecided | R2-5, R2-6
R1-26 | closed_with_regression | wrapper start/end snapshots to operator-owned log specified; no reader/comparison defined | R2-7
R1-27 | closed_with_regression | per-rung gate-survival table added, rung-3/4 adoption graduation-grade; rung-0 cells overstate | R2-16
R1-28 | closed_with_regression | step-0 denial canary added, marker-loss fail-closed; canary mechanism unspecified AND cannot isolate the fence layer | R2-1, R2-2
R1-29 | closed_with_regression | resume cap k=3 + DEAD marker + doctor dead-man line added; pull-only reader + per-dir-not-per-cause bound | R2-9, R2-10
R1-30 | closed | mirror line count corrected to 1,557; merge grep clean (L1 HIGH) | —
R2-1 | closed_with_regression | canary two-phase stream-json drive + positive fired-record specified, flag pair leaf-verified (L2 HIGH); undocumented mid-drive --json-schema leg + unprobed behavioral legs remain | R3-1
R2-2 | closed | positive fired-record hook-liveness (nonce+decision=deny, aborts on marker loss) replaces the deny-outcome canary; smoke test verifies fence-dormant abort; lead-directed fix (L6 sound) | —
R2-3 | closed | FEOV execution locus determined by probe: session-Bash setup/capture + Workflow-tool debate.js (hook-scoped scriptPath); MIXED-locus statement leaf-verified against pinned research.md (L2 HIGH) | —
R2-4 | closed | §2.2 step 2 reworded to the wrapper-side architecture (L5 spot-verified in place) | —
R2-5 | closed_with_regression | causal-provenance origin tagging (wrapper stamps .sleeper-origin at creation + ledger-window fallback); voids at rung 0 / dead-run path / red-memory surface | R3-2, R3-3, R3-4
R2-6 | closed_with_regression | disjunction decided → severity-gated bypass (infra classes flagged sleeper-only, ordinary need corroboration); classifier source unstated + doctor-channel overclaim + rung-0 void | R3-2, R3-5, R3-6
R2-7 | closed_with_regression | wrapper compares start/end + start-vs-prior-end guardrail hashes → TAMPER flag + fail-closed preflight; abort path uncovered + out-of-repo write blind spot | R3-7, R3-8
R2-8 | closed | preflight recomputes plugin-copy hash, fails closed on mismatch; refresh in scheduling.md checklist + doctor staleness line (L6 sound; direct read line 611) | —
R2-9 | closed | dead-man made push via SessionStart-hook warning in every interactive session (invariant 7 human surface); lead-directed (direct read line 745) | —
R2-10 | closed_with_regression | per-cause dead SIGNATURE + M=3 HALT + softened claim added; signature normalization unspecified (never fires) + burn arithmetic wrong | R3-11, R3-12
R2-11 | closed_with_regression | graduation-queued status exempts from 30-day auto-stale while still deduping; exemption has no upper bound (permanent suppression) | R3-13
R2-12 | closed | idempotence completeness derived from the wrapper's own step-7 operator-owned ledger record, not loop-writable state (L6 sound) | —
R2-13 | closed | §6 row 5 cell requalified "(no spend-limit API; rate-limit API unreachable at this auth tier — §5.1/R1-9)" (direct read line 1122) | —
R2-14 | closed | §7 Pattern B/E bullet appended "(upgraded to leaf-verified HIGH round 1, R1-11 — lag fixed round 2)" + R1-11 added to §7 upgrade list (L4 verified) | —
R2-15 | closed | §3.4 rung-1 label qualified "RECOMMENDED default AMONG SCHEDULED RUNGS" (L5 spot-verified) | —
R2-16 | closed_with_regression | R0 L2 cell split (fence YES cache copy; canary n/a) + manual-spend out-of-ledger stated; rung-0 execution shape for steps 0/2/4/7 still undefined | R3-2
R2-17 | closed | §3.3 picks the trade: strict-mcp-config qmd-only, round-1 ToolSearch pdf/arxiv parenthetical corrected as false under the flag (direct read lines 676-680) | —
R2-18 | closed | expected per-run spend owned; ceiling-vs-cap composition stated as intended month-end throttle / anomaly signal; arithmetic recomputed (L1/L5) | —
R2-19 | closed_with_regression | §0 enumeration made total over the round-1 tree (skill file + manifest added); round-2 minted artifacts fall outside it | R3-10
R2-20 | closed | [^Backlog] pin range → 15–18 (L5 spot-verified) | —
R2-21 | closed | DGM "exact"→"direct" + honesty clause (evaluates before archive, admits low scorers; our gate stricter); arxiv/html re-fetched live (L2 HIGH) | —
R2-22 | closed | [^EfficiencyPlan] added beside the run-3 $149.95 figure (L5 spot-verified) | —
R3-1 | closed_with_regression | canary --json-schema leg demoted to verify-at-build + fenced-block fallback + OQ22/OQ23; degrade-note readers unspecified at both reader sites | R4-7
R3-2 | closed_with_regression | rung 0 invokes SAME wrapper via --manual ("same code path" true by construction, out-of-ledger dissolves); /self-improve command-invocation surface + inverted containment polarity + missing disable-model-invocation remain | R4-1
R3-3 | closed_with_regression | provenance tagging extended to red-memory mirror (sleeper-authored excluded); mechanism aimed at a profile-foreclosed write + change-keyed granularity over-tags extended human patterns | R4-4
R3-4 | closed_with_regression | run-window START logged at step 0 + DEAD-time origin sweep; unobserved-exit degenerate case leaves multi-day window that sweeps human dirs; unclosed-window read unspecified | R4-6
R3-5 | closed | infrastructure-class tag assigned SOLELY from the wrapper's own event log; friction TEXT never self-classifies (L1 verbatim, CLEAN) | —
R3-6 | closed | doctor line prints per-signature counts since last clear, keyed by R3-11 normalization (one mechanism fixes R3-6+R3-11; L1 CONSISTENT) | —
R3-7 | closed | wrapper snapshots at EVERY observed exit, compare chained to last-recorded-snapshot-of-any-kind, missing link fails closed (L5 spot-verified; crash-without-end covered by START snapshot) | —
R3-8 | closed | leg (b) scoped honestly to in-repo/guardrail sensing; out-of-repo residual bounded by pinned code + no-write-gadget (post-R3-15); body complete (CHANGELOG bullet omitted → R4-16) | —
R3-9 | closed_with_regression | cheap parts absorbed (watchmen in layer 3 + snapshot set, absolute-path invocation, last-check timestamp) + self-verification limit owned; dead-man disarm flag custody + R3-9/R3-10 same-line composition unstated | R4-14
R3-10 | closed_with_regression | §0 enumeration extended (SessionStart hook + cross-plugin doctor delta + operator configs), warning conditioned on scheduling-enabled; "exactly THREE code artifacts" count headline unreconciled with the fourth executable | R4-8
R3-11 | closed | dead SIGNATURE normalization specified (template, dates/paths/ids/nonces/digits stripped), zero-HALT-firings telemetry + alternating-cause residual owned (L1/L2/L5 CONSISTENT) | —
R3-12 | closed_with_regression | retry semantics stated + recomputed (≤$5/night, cap ~10 nights); cap-trip treated as terminal but the cap RESETS — HALT lands early month 2, worst-case ≈ one cap + ε | R4-10
R3-13 | closed_with_regression | graduation-queued gains M-day queued-stale re-surface; "no status is timer-free" false — rejected/graduated have no timer/dedupe semantics (missing root invariant, 3rd per-status patch) | R4-9
R3-14 | closed_with_regression | §3.2 dontAsk carve-out corrected + carve-out deny-enumeration + belt denies; enumeration rests on non-exhaustive doc list (R4-3) + prior-exposure example refuted by deny-reach clause (R4-5) | R4-3, R4-5
R3-15 | closed_with_regression | git --output belt denies + hook matcher (3 flags) + OQ18 scope; sibling git-native write flags (-o/format-patch/archive/bundle) outside both (R4-2) + probe-attribution layer-masked (R4-11) | R4-2, R4-11
R3-16 | closed | §1.3 telemetry row updated "SHIPPED as of FEOV 0.7.0 — present in this run's own trajectories/" (L1 filesystem-verified, HIGH) | —
R3-17 | closed | [^Pricing] tokenizer set completed with Opus 4.7+; live re-fetch verbatim (L3 HIGH) | —
R4-1 | closed_with_regression | thin disable-model-invocation trampoline + payload out of commands/ + §3.4 polarity corrected; ladder row 0 + §3.3 Phase-4 acceptance test still specify the OLD shape, and the paste-run pool residual's deferral never argues the rung-2 case | R5-2, R5-3
R4-2 | closed_with_regression | git channel inverted to hook read-ALLOWLIST + belt extended to -o/-O + writer family; "Where Bash IS reachable (the Workflow seat agents…)" contradicts layer 4 (iii), and the ATTACHED -o<value> form escapes the new belt (merge-reproduced) | R5-1, R5-5
R4-3 | closed_with_regression | bare `Bash` deny closes the carve-out class for the TOP-LEVEL session (doc-verified, sound); scope over the Workflow seat agents is contradictory/unestablished — settings inheritance never probed | R5-1
R4-4 | closed_with_regression | denied-by-design horn picked + stated plainly + window-ADDED granularity; the expected-memory-deny bucket is TARGET-keyed, laundering attack-attempt evidence on the memory surface | R5-6
R4-5 | closed | postmortem corrected at BOTH sites (§4.2 + §6 row 13), example re-pointed at credentials-class paths, deny-reach clause noted (L3 live-doc verified faithful) | —
R4-6 | closed_with_regression | window END bounded by next START + retroactive-uncertain flag + confinement; confinement is dir-NAME-keyed against §1.5's own doctrine and "the wrapper's own sub-run slug" is unknowable after a hard-kill | R5-4, R5-3
R4-7 | closed | §2.3 confidence carries the qmd-degrade labeling clause; §3.4 doctor carries the degrade-streak term; both reader sites now obligated (L2/L5 verified in place) | —
R4-8 | closed_with_regression | count corrected to FOUR + host plugin named sleeper-service; §0 TREE never reconciled (no hooks.json/SessionStart entry) + new hooks.json bootstrap-guard coverage unstated | R5-8, R5-7
R4-9 | closed_with_regression | root invariant stated once in §1.4, terminal states given semantics, §2.3/§6 row 3 carry it; both `rejected` arms key on an unrecorded rejection date + `regression` token domain unpinned | R5-9
R4-10 | closed | cap/HALT arithmetic recomputed correct — independently re-walked by L2, L5, AND L6 (deaths nights 4/8; cap-skip from ~night 11; HALT ~$5–10 into month 2; ≈$55–60 across two months) | —
R4-11 | closed | attribution re-scoped "consistent with … does NOT isolate" at §4.2, §7, and OQ23(d); isolating dontAsk-zero-allow probe deferred to build (L5 verified all three sites) | —
R4-12 | closed_with_regression | est_complexity default-1-inert + backlog-note source stated; note FORMAT unstated and the named source is unpopulated in the pinned corpus (L1 leaf: no parseable field on any of 25 items) | R5-10
R4-13 | closed_with_regression | provenance/corroboration gate row added (YES R0/R1, NO R2–R4) + named in the adoption requirement; the row is the third per-surface pool patch — the pool root invariant is still missing and rung-2 poisoning is automatic behind an instructional gate | R5-3
R4-14 | closed_with_regression | flag custody operator-owned + never-fully-silent three-state print (disabled/missing/stale distinguishable); flag-ABSENT-as-anomaly nags every never-configured install (first-install default state) | R5-7
R4-15 | closed | OQ18(c) extended to the git SUBCOMMAND boundary (config/gc/repack/maintenance probe cases); hook read-allowlist rejects them by default (L5/L3 verified) | —
R4-16 | closed | CHANGELOG Round 3 now carries the R3-8 bullet (L5 verified; change-summary channel reconciled) | —

## NOTES — upgrades blue may bank / non-gaps (not open gaps)

- **#68375 volatility signal (L4, round 5 — not a gap):** the issue now carries a GitHub
  `stale` label beside `regression`/`has repro`. Still OPEN (zero drift on content), but a
  bot auto-close is a live drift risk; keep re-checking. Row-6-style "open ≠ will-be-fixed"
  framing unaffected either way.
- **[^MissedRun] re-point suggestion (L2, round 5):** the r1-era learn.microsoft.com Task
  Scheduler settings URL now 404s; the quoted MMC checkbox string has no stable doc leaf.
  Semantics re-verified at the TaskSettings.StartWhenAvailable API property page ("the Task
  Scheduler can start the task at any time after its scheduled time has passed") — the
  footnote should cite that page. Mechanism HIGH; exact-UI-string leaf MEDIUM. No gap:
  design conclusion untouched.
- **§7 round-4 update bullet absent (L5, sub-trivial, not minted):** rounds 1/2/3 each got
  a §7 self-audit bullet; round 4's leaf verification (format-patch probe) lives only in
  §4.2/CHANGELOG. Copy edit if blue touches §7 anyway.
- **OQ17 answerable at leaf today (standing since r4):** `permissions.disableAutoMode`
  documented in the same sentence as the bypass lockout; blue keeping OQ17 open is
  conservative, not a defect. Optional, trivial.
- **[^AlertFatigue] replacement figure (standing since r3, unbanked):** pinnable 2026
  survey figure ("57% report fewer than 30% of alerts actionable," n=1,039) + ACM Computing
  Surveys doi 10.1145/3723158. Optional, trivial.
- **[^PortPlan]** remains snapshot-grade; pin-absent defect is R1-7, LEAD-adjudicated.
  AgentOrange working tree re-confirmed 6df52af-clean round 5 (L1/L2), quotes verbatim,
  zero drift over 3 rounds.
- **[^HeadlessProbe] P1/P2 figures MEDIUM:** ephemeral instrument;
  disposition-of-record stands (re-run + commit at build).
- **Volatile leaves re-fetched zero-drift r5:** [^ScheduledTasks] (3 claim sets),
  [^RoutinesDocs] (full rung-3 set), [^MissedRun] systemd+anacron, [^GhaSchedule],
  [^WebSandbox], [^PermissionsDoc] (full §4 quote set incl. bare-name removal + deny-reach
  + carve-out include-list), [^HooksDoc], [^Pricing] full grid, [^IdeaStudy], #76239/#68375
  OPEN (5th round), #837/#14246/#22055/#6631/#25621 statuses, R4-2 gadget re-reproduced
  (spaced AND attached forms, L3/L4/L6 + merge).
- **Invariant 8** stands; R5-1 is the residual where "every execution mode/surface" was
  argued over the top-level session's steps but not derived over the seat population the
  design's own step 4 spawns.

---

### red/archive.md (verbatim)

---

# red/archive.md — closed-gap prose record (APPEND-ONLY; never edit an existing block)

Immutable full-prose records of closed gaps: what was found, how verified, closure class.
The ledger's closure index is the screen; this file is the evidence of record.

(no closed gaps yet — round 1)

## Round 2 closures (red merge, 2026-07-17)

### R1-1 — backlog item count fails recount (40 → 25) — CLOSED
Found round 1 (L1): §1.3/[^IdeasCorpus] claimed "40 statused items"; recount at pin
`7bc501e` = 25 checkbox items / 39 lines. Round-1 repair: body + footnote restated
"25 statused checkbox items across 39 lines ... recounted at the pin, R1-1". Verified
round 2 by L1 re-count via `git show` (25/39 reproduced) and merge grep: "40 statused"
survives only inside §7's propagation-grep token log. Closure class: **closed**.

### R1-2 — scope-fusion overstatement on the PDF-gap recurrence claim — CLOSED
Found round 1 (L1): fused "red, blue, AND judge across two consecutive runs". Repair
splits the sentence into backlog 27c (three seats, one run) and 31h (two runs, red
merges), labeled "two sources, stated separately (round-1 correction R1-2)". Verified
round 2 by L1 at the pin: line 27(c) and line 31(h) verbatim; line-scheme validated by
two independent anchors. Closure class: **closed**.

### R1-3 — [^DGM] homes a Sakana-post quote to the arXiv paper — CLOSED
Repair: "improve themselves the more compute they are provided" moved to [^DGMSakana];
"(ICLR 2026)" dropped; [^DGM] carries the correction note. Verified round 2 by L1
(footnote text inspected; r1 leaf reads of record: quote verbatim at sakana.ai/dgm,
absent from abs and /html). §1.2's dual-cite [^DGM][^DGMSakana] acceptable — the
carrying source is cited. Closure class: **closed**.

### R1-4 — [^SICA] venue metadata contradicted the cited page — CLOSED
Repair: venue restated per the abs page's Comments field ("Submitted as a preprint to
NeurIPS 2025"); workshop tag dropped. Verified round 2 by L1 live re-fetch: Comments
verbatim; 17–53% SWE-bench-Verified-subset figure re-confirmed. Closure class: **closed**.

### R1-5 — "three MCP issues leaf-checked OPEN" while #32191 was CLOSED-duplicate — CLOSED
The sharpest round-1 citation defect (false verification claim). Repair: §3.3 restates
"TWO leaf-checked OPEN (#76239, #68375) ... ONE historical: #32191 ... CLOSED as
duplicate (canonical untraced; 2.1.58–2.1.71 era)"; §6 row 8 and [^McpHeadlessBugs]
propagated. Verified round 2: L2 confirms the restatement matches ledger statuses; L4
re-fetched both OPEN issues live (drift check — both still OPEN 2026-07-17); merge grep
confirms no stale "leaf-checked OPEN" claim covering #32191 remains. Closure class:
**closed**.

### R1-6 — exit code "0/1" attributed to a page with no exit-code docs — CLOSED
Repair: §3.2 requalified — 0-on-success probe-corroborated; "on error the CLI exits
nonzero, but the cli-reference page publishes no exit-code table ... the wrapper treats
ANY nonzero exit as failure"; [^CliReference] carries the correction. Verified round 2 by
L2 on a fresh cli-reference fetch: still no exit-code table; `--max-turns` still says
only "Exits with an error." Closure class: **closed**.

### R1-8 — disableBypassPermissionsMode misplaced at top level (silent no-op) — CLOSED
Repair: key moved INSIDE the `permissions` object in the §4.2 sample (line "INSIDE
permissions — R1-8"), with prose explaining the silent-ignore failure. Verified at the
round-2 merge by direct read of the sample JSON against the r1 citation-ledger doc line
("permissions.disableBypassPermissionsMode", any scope). `disableAutoMode` correctly
deferred to OQ17 pending leaf-verify rather than asserted. Closure class: **closed**.

### R1-9 — "no endpoint to read" stale vs the live Rate Limits API — CLOSED WITH REGRESSION
Repair: §5.1 + H5 verdict requalified (spend limits: no API read/set; rate limits:
API-readable, read-only, Admin-key-only — nothing a subscription-auth scheduler can
poll); [^ConsoleLimits] amended; [^RateLimitsAPI] added. Verified round 1 at the leaf by
L3; round 2 L3 found the §6 row 5 likelihood cell still reads the pre-repair flat
"(no API)". Closure class: **closed_with_regression** → successor **R2-13**.

### R1-10 — [^CliReference] mislabeled --fallback-model print-only — CLOSED
Repair: label dropped; §5.1 bullet and footnote note the persistent `fallbackModel`
setting. Verified against the r1 leaf fetch (L3); no round-2 contradiction on the
re-fetched cli-reference page. Closure class: **closed**.

### R1-11 — "frontier ~$10/$50" true only of Fable/Mythos class — CLOSED WITH REGRESSION
Repair: §5.2 + [^Pricing] pinned to leaf figures, two frontier tiers named, tokenizer
+30% caveat added, Batch ≤24h honestly carried MEDIUM. Verified round 2 by L3 live
re-fetch: zero drift on all figures; the ≤24h sub-claim independently resolved HIGH on
the batch-processing page (bankable). Regression: §7's self-audit still says "pricing
figures graded MEDIUM" and its upgrade list omits R1-11 (L4). Closure class:
**closed_with_regression** → successor **R2-14**.

### R1-12 — bare-relative deny rules anchor at process cwd — CLOSED
Repair: all guardrail rules written `//`-absolute in the sample; anchoring physics
paragraph added (bare = cwd, `/` = settings source, `//` = root); wrapper-cd precondition
stated for scheduling.md beside the trust-dialog one. Verified at the round-2 merge by
direct read of §4.2 against the r1 citation-ledger anchor-table line. Closure class:
**closed**.

### R1-13 — chmod-444 dressed as a documented community pattern — CLOSED
Repair: [^PermAskBypass] splits the claims — PreToolUse exit-2 protected-files workaround
kept as thread-verbatim; chmod-readonly recast as "a DESIGN-PROPOSED measure of this
report's own, not community-sourced" in §4.3 layer 3. Verified at the round-2 merge by
direct read of layer 3 + footnote against the r1 full-thread gh check (L4/L3 concurring).
Closure class: **closed**.

### R1-14 — marginal value over the manual process never argued — CLOSED WITH REGRESSION
Repair: §1.4 gained the priced null-alternative paragraph — what automation buys (bounded
research + structured stubs, mechanical recurrence arithmetic, no single-point recall),
what it costs, rung 0 as DEFAULT and possibly terminal, cadence as hypothesis with a
named triage-rate revisit trigger. Verified by full read; the demanded honest paragraph
exists and engages the report's own Dependabot evidence. Regression: §3.4's ladder still
stamps rung 1 "RECOMMENDED default" unqualified (L5). Closure class:
**closed_with_regression** → successor **R2-15**.

### R1-15 — preferred --bare recipe made the repo copy the executing copy — CLOSED WITH REGRESSION
Repair: `--plugin-dir` pinned to an operator-owned, read-only copy outside the write
surface, "NEVER the repo's `plugins/` working tree"; §4.3 layer 2's route-around updated.
Verified by full read (§3.2 + layer 2 consistent). Regression: the new copy's lifecycle
is unbuilt — no creation/refresh owner, "the human-gated plugin update path" defined
nowhere, preflight asserts load not version (L5+L6 independently). Closure class:
**closed_with_regression** → successor **R2-8**.

### R1-16 — Bash(node scripts/*): write-capable channel, layout mismatch — CLOSED WITH REGRESSION
Repair: the allow rule deleted; harvest.mjs moved wrapper-side pre-session; session Bash
= pinned-argv read-only git; compound/redirection/traversal matching named as build-PR
test (OQ18). Verified by direct read of the §4.2 sample + §1.3 + merge grep (`Bash(node`
survives only in correction contexts). Regressions: §2.2 step 2 still reads "SCORE
(script...)" in-session (L5); step 4's FEOV invocation never reconciled with "removes ALL
script execution from the session" — the load-bearing composition defect of the round
(L5). Closure class: **closed_with_regression** → successors **R2-3, R2-4**.

### R1-17 — read/exfiltration channel never considered — CLOSED
Repair: Read/Grep/Glob allow-scoped to repo + plugin copy; belt Read-denies on
`~/.claude` settings/credentials/session transcripts; WebSearch dropped from the nightly
profile; §6 row 13 added with an argued RISK-ACCEPT (single-operator box, largely-public
repo, narrow enumerable session-logged residual egress). Verified by direct read of the
profile JSON + row 13. The acceptance is argued as demanded — red accepts the argument.
Closure class: **closed**.

### R1-18 — injection via the retrieval surface had no risk row — CLOSED
Repair: §6 row 14 added (graded, disposed); stub contract gained `origin-labels`
(web-derived = untrusted-origin, barred from ranking inputs); WebSearch nightly drop
shared with R1-17; §2.2 step 4 names the narrowed nightly retrieval surface. Verified by
direct read. Closure class: **closed**.

### R1-19 — monthly ledger inside the loop's write surface (invariant 4) — CLOSED WITH REGRESSION
Repair: ledger relocated to `~/.claude/sleeper/ledger.jsonl`, wrapper-written outside the
session, fail-closed on unparseable ledger; invariant-4 layer for the monthly guard named
(the wrapper); §0 invariant 5 updated. Verified by direct read of §5.1 + §2.2 step 7.
Regression: the gap's own text named the idempotency-gate sibling ("Same shape, lower
stakes"), which got neither fix nor argued acceptance — completeness signal and DEAD
marker location remain loop-writable/unstated (L5). Closure class:
**closed_with_regression** → successor **R2-12**.

### R1-20 — H3 "Confidence HIGH" attached to the untested bare path — CLOSED
Repair: verdict split — "HIGH for the probed non-bare path; OPEN for the preferred
`--bare --plugin-dir` recipe", with "A verdict stamp must not grade higher than its
recommended configuration's evidence." Verified round 2 by L2 (CLEAN). Closure class:
**closed**.

### R1-21 — row 4 accepted on the benign-actor premise §4.1 rejects — CLOSED WITH REGRESSION
Repair: row 4 re-argued explicitly WITHOUT actor intent — legs: (a) channel narrowness
(pinned git argv only), (b) detection outside the model's reach (R1-26 snapshots),
(c) L×I×Cx vs Windows sandbox complexity; "No premise about actor intent is used."
Verified by direct read — the re-argument is performed as demanded. Regression: leg (a)'s
premise is contested by the unreconciled step-4 FEOV execution locus (if a workflow
runner executes scripts outside the Bash surface, "the only subprocess reachable is the
pinned git set" is false), and leg (b) leans on a snapshot channel with no defined reader
(see R2-7). Closure class: **closed_with_regression** → successor **R2-3** (carries the
re-argument obligation; R2-7 covers leg (b)).

### R1-22 — "stubs age out visibly" was policy-without-mechanism — CLOSED WITH REGRESSION
Repair: aging specified — skip only stubs younger than 30 days; older untriaged stubs
auto-stale via harvest's filename-date sweep and re-enter the docket flagged
`stub-stale`; §2.3 status field carries the lifecycle; §6 row 3 updated. Verified by
direct read of §1.4 stage 2 + §2.3 + row 3. Regression: the window composes badly with
human graduation latency — a graduation-queued good stub is re-researched every 30 days
(L6). Closure class: **closed_with_regression** → successor **R2-11**.

### R1-23 — Batch API as a lever with no routing mechanism — CLOSED
Repair: demoted to an explicit FUTURE note ("if and when a routing mechanism exists, and
not before"). Verified by direct read of §5.2. Round 2 bonus: the ≤24h window sub-claim
resolved HIGH at the batch-processing page (L3). Closure class: **closed**.

### R1-24 — layer-6 row asserted what OQ3 admits is unverified — CLOSED
Repair: the qualifier carried inline in the table row ("doc-verified for SCHEDULED fires;
the in-session hostile-invocation case is exactly open question 3"), fallback layers
named in the route-around cell. Verified by direct read of §4.3 layer 6. Closure class:
**closed**.

### R1-25 — the loop authors its own future harvest inputs — CLOSED WITH REGRESSION
Repair: §1.5 extended to authorship — harvest tags `*_self-improve/` rows
`origin: sleeper`, caps sleeper-origin recurrence at 1 per class, requires non-sleeper
corroboration (or a `sleeper-only` flag) before the pick; §6 row 10 re-scoped and
re-graded; §4.3's "eats its own complaints" now carries the cap. Verified by direct read.
Regressions: (1) the tag is keyed on a dir-NAME glob and the loop's own spawned FEOV
sub-run dir escapes it — worse, the untagged sub-run friction satisfies the
"non-sleeper occurrence" gate, self-supplying the demanded corroboration (L6); (2) the
"requires X — or, failing that, flags" disjunction is undecided between a hard bar and a
labeled pass, and each horn has an unowned cost (L5+L6). Closure class:
**closed_with_regression** → successors **R2-5, R2-6**.

### R1-26 — uncommitted-working-tree tamper evidence was self-erasing — CLOSED WITH REGRESSION
Repair: wrapper takes out-of-session snapshots (git porcelain + SHA-256 guardrail hashes)
at run start and end, appended to an operator-owned log outside the repo; §4.3 layer 5
restates the durable-tamper-evidence property; heavier auto-commit alternative named.
Verified by direct read of layer 5 + §2.2 steps 0/7. Regression: the snapshot log has no
defined reader or comparison step — the same undefined-receiver defect R1-29 fixed for
loop death, unfixed here; row 4's leg (b) presupposes the missing watchman (L5+L6).
Closure class: **closed_with_regression** → successor **R2-7**.

### R1-27 — gate stack sold as rung-portable — CLOSED WITH REGRESSION
Repair: per-rung gate-survival table added to §3.4 (6 controls × 5 rungs), rung-4
settings-inversion named, "climbing past rung 2 is itself a graduation-grade decision —
never a config toggle"; §6 row 16 added. Verified by direct read; L2 cross-checked the
table's cells against ledger-verified [^RoutinesDocs]/[^ScheduledTasks] facts — no cell
contradicts a verified source. Regression: the rung-0 column — the design's own
newly-default mode — holds the soft cells: compound L2 row claims YES for a canary that
needs the absent wrapper, and manual-run spend never enters the monthly ledger, unstated
(L5). Closure class: **closed_with_regression** → successor **R2-16**.

### R1-28 — write-fence marker-keyed, fails OPEN on marker loss — CLOSED WITH REGRESSION
Repair: step-0 denial canary added (attempt one out-of-fence write, abort unless
DENIED); §0 tree, §2.2 step 0, §4.3 layer 2 all carry it; OQ2 reframed
verify-enforcement-not-presence; OQ16 (polarity inversion) added. Verified by direct
read. Regressions — the canary as specified does not deliver the closure it stamps:
(1) actor/observer/abort mechanism unspecified, with the step-0 header ("wrapper,
OUTSIDE the model session") contradicting the "First in-session action" label — as
printed it is either a post-hoc check or an instructional abort of the class §4.1
refutes (L5); (2) layers 1 and 2 fence the identical boundary, so a fence-dormant run
with live permissions PASSES the canary — it cannot isolate the layer it certifies, and
OQ2's own acceptance test fails as designed (L5+L6 independently). Closure class:
**closed_with_regression** → successors **R2-1, R2-2**.

### R1-29 — resume-forever livelock + undefined receiver of "loudly" — CLOSED WITH REGRESSION
Repair: resume cap k=3 → DEAD marker + dated abort record + fresh dir; dead-man surface
defined (wrapper-maintained `last-successful-run`; doctor reports "last successful
sleeper run: N days ago"); §6 row 15 added. Verified by direct read of §3.4 + row 15.
Regressions: (1) the reader is pull-only and "a surface the human already looks at" is
asserted, not evidenced — the automation's purpose removes exactly that habit (L6);
(2) the cap bounds per-DIR, not per-CAUSE — a deterministic wedge burns ~k×$5 nightly
via fresh dirs until the monthly cap trips, resetting monthly (L6). Closure class:
**closed_with_regression** → successors **R2-9, R2-10**.

### R1-30 — gap-pattern mirror line count off by one — CLOSED
Repair: corrected to 1,557 in §1.3 and [^RedPatterns] with the byte-exact recount note
(final byte 0x0a). Verified round 2 by L1 (`wc`: 1,557 lines / 119,418 bytes) and merge
grep: the only surviving "1,558" is §7's propagation-grep token log. Closure class:
**closed**.


## Round 3 closures (red merge, 2026-07-17)

### R1-7 — inputs/PINNED.md pins a path absent at the pin — ADJUDICATED (risk_argued)
Standing since round 1; carried with the LEAD through round 2. LEAD ruling (debate.md round-2
LEAD section): **risk_accepted**. The finding is valid and verified (`git cat-file -e` MISSING for
`plans/claude-port-plan.md` at `7bc501e`), but the fix (setup-script pin validation / staging
the port plan into inputs/) is run-infrastructure owned by the lead, not report text blue can
change; the affected quotes were re-verified verbatim at working-tree `6df52af` (r2 L1),
re-confirmed 6df52af-clean this session (r3 L1), so residual impact is a reproducibility
caveat already disclosed in [^PortPlan]. Recorded, not dropped: port-plan citations REMAIN
snapshot-grade in the final report; the lead owes pin validation before the next run.
Excluded from red's verdict pool per the run directive. Closure class: **risk_argued**.

### R2-1 — canary actor/observer/abort mechanism unspecified — CLOSED WITH REGRESSION
Found round 2 (L5, superseding R1-28): the step-0 canary had no mechanism connecting the
in-session write attempt to the wrapper's abort, and the header/label contradicted. Round-3
repair: blue specified the two-phase stream-json drive (`--input-format stream-json
--output-format stream-json`, flag pair leaf-verified by L2 in `claude --help` on CLI 2.1.212,
HIGH) plus a positive fired-record — the wrapper drives a canary prompt, parses the deny
event, confirms the sleeper-guard fired-record carries the nonce with decision=deny, only then
prompts real work. Verified round 3 (L2 flag HIGH; L6 mechanism sound). Regression: the pick
returns "as structured output (the phase drive's `--json-schema` leg)" — an undocumented
mid-drive schema-bound composition (structured output is documented only for `--output-format
json` final results, L2); and three behavioral legs of the drive plus two invariant-7 staging
edges are unprobed/unstated (L5). Closure class: **closed_with_regression** -> successor **R3-1**.

### R2-2 — the deny-outcome canary cannot isolate the hook fence — CLOSED
Found round 2 (L5+L6, superseding R1-28): layers 1/2 fence the identical boundary, so a
fence-dormant run passes the deny-outcome canary. Round-3 repair implements exactly the LEAD's
directed fix: a POSITIVE hook-liveness record — the sleeper-guard hook writes a fired-record
(nonce + decision=deny) that the wrapper confirms non-empty per run; marker loss -> no
fired-record -> ABORT (the H4-refuted permissions-only configuration now fails closed). The
smoke test (report lines 1211-1213) verifies the fence-dormant run aborts as designed —
resolving the R2-2 contradiction between OQ2's acceptance test and the claim. Verified round 3
by direct read (§4.3 layer 2, lines 424-427) + L6. Closure class: **closed**.

### R2-3 — step-4 FEOV locus never reconciled with "removes ALL script execution" — CLOSED
The round-2 HIGH (L5, superseding R1-16, R1-21). Round-3 repair: blue determined the FEOV
execution locus BY PROBE and stated the MIXED shape — session-Bash `node setup-research-run.mjs`
(step 2) and `node capture-research-run.mjs` (step 5), with `debate.js` executed via the
**Workflow** tool (harness-side runner, hook-scoped `scriptPath`), outside the Bash permission
surface. Verified round 3 (L2): read the pinned `research.md` (shipped 0.7.0 copy AND
`git show 7bc501e:` byte-identical) — step 2/3/5 loci exact; `--smoke` = 1 lane/1 round/haiku/
~50k tokens; capture emits cost.md + run-record-audit.md — the mixed-locus statement is exact,
HIGH. L6 concurs sound. The row-4 "read-only git argv" premise is separately challenged by the
`git log --output` write gadget — a NEW finding (R3-15), not an R2-3 residual. Closure class:
**closed**.

### R2-4 — §2.2 step 2 still executed a script in-session — CLOSED
Found round 2 (L5, superseding R1-16): propagation miss. Round-3 repair: step 2 reworded to
the wrapper-side architecture (read the staged scored table / model-side language-tier scoring,
no session script). Verified round 3 by L5 spot-check in place. Closure class: **closed**.

### R2-5 — authorship cap keyed on a dir-NAME glob — CLOSED WITH REGRESSION
Found round 2 (L6, superseding R1-25): the loop's spawned sub-run landed untagged friction that
satisfied the corroboration gate. Round-3 repair: §1.5 rewritten to tag by CAUSAL PROVENANCE —
the wrapper stamps `inputs/.sleeper-origin` (carrying the parent run id) into EVERY run dir it
creates, and harvest additionally treats any dir whose creation timestamp falls inside a
wrapper-logged sleeper window as sleeper-origin even if markerless. Verified round 3 (L1
leaf-read `setup-research-run.mjs`: the wrapper choosing runDir + stamping is buildable as
claimed, HIGH). Regressions (three surfaces the marker/dir keying does not reach):
(1) at rung 0 — the DESIGN'S DEFAULT mode — there is no wrapper, so both manual run dirs land
markerless and outside any ledger window (L5 -> R3-2, jointly superseding R2-6, R2-16);
(2) the red gap-pattern MEMORY mirror is a separate surface with no run dir, so nightly-authored
patterns enter harvest un-tagged as "non-sleeper" (L6 -> R3-3); (3) the ledger-window fallback is
void for DEAD runs (no step-7 ledger write -> no logged window -> no porcelain sweep), letting a
model-authored in-session dir on a dying night escape markerless (L1 -> R3-4). Closure class:
**closed_with_regression** -> successors **R3-2, R3-3, R3-4**.

### R2-6 — undecided "requires corroboration — or, failing that, flags" disjunction — CLOSED WITH REGRESSION
Found round 2 (L5+L6, superseding R1-25). Round-3 repair: blue decided the branch — a
severity-gated bypass (infrastructure-failure classes enter flagged `sleeper-only` WITHOUT
corroboration; ordinary classes require a non-sleeper occurrence), the shape the LEAD endorsed
as a starting point. Verified round 3 by direct read (§1.5) + L6 (correct decision, owns both
horns). Regressions: (1) the report never states WHO classifies a row as infrastructure — the
only specified classifier is keyword clustering over friction TEXT, so loop-authored/injected
friction can self-route into the no-corroboration lane (L6 -> R3-5); (2) "every infrastructure
class ALSO surfaces independently on the doctor line" overclaims — the line prints only the
last skip reason, so transient events older than the most recent are on no channel but the stub
flag (L1+L5 -> R3-6); (3) the whole gate is void in rung-0 default mode (L5 -> R3-2). Closure
class: **closed_with_regression** -> successors **R3-2, R3-5, R3-6**.

### R2-7 — wrapper start/end snapshot had no reader or comparison step — CLOSED WITH REGRESSION
Found round 2 (L5+L6, superseding R1-26). Round-3 repair: the wrapper now COMPARES end-vs-start
and start-vs-prior-run's-end guardrail hashes at step 7; on mismatch it writes a dated TAMPER
flag beside the ledger and the next preflight fails closed (report line 479-483). Verified round
3 by direct read + L5/L6 (the watchman exists and is sound — not reopened). Regressions:
(1) the compare runs ONLY at step 7, so a run dying before it takes no END snapshot and
"start-vs-prior-END" has no referent — the tamper chain has an unspecified link at exactly the
abort path, which is model-influenceable (L5 -> R3-7); (2) the sensing surface (`git status
--porcelain` + guardrail-file hashes) is blind to out-of-repo subprocess writes, which is where
the layer-4 debate.js residual (leg a concedes it writes files) can land (L6 -> R3-8). Closure
class: **closed_with_regression** -> successors **R3-7, R3-8**.

### R2-8 — operator-owned plugin copy had no refresh owner or version check — CLOSED
Found round 2 (L5+L6, superseding R1-15). Round-3 repair: the wrapper preflight recomputes the
copy's content hash and FAILS CLOSED on mismatch with the operator-approved value (report line
611); the refresh step is added to scheduling.md's guardrail-PR merge checklist; a doctor line
reports copy staleness. Verified round 3 by direct read (line 611) + L6 (sound). The remaining
question — whether the wrapper VERIFYING the copy is itself executing from that copy
(self-verification) — is a NEW reflexivity finding about the wrapper's own root of trust
(R3-9), not an R2-8 residual. Closure class: **closed**.

### R2-9 — dead-man surface was pull-only — CLOSED
Found round 2 (L6, superseding R1-29). Round-3 repair: the staleness check now runs as a
SessionStart-hook warning line in every interactive Claude Code session (report line 745) — a
push surface the operator receives passively "whenever they use Claude Code at all" (invariant
7's human surface), directed by the LEAD. Verified round 3 by direct read. The warning's
enablement condition (an operator who schedules then stops leaves it firing forever) is folded
into the enumeration/hygiene successor R3-10, not carried here. Closure class: **closed**.

### R2-10 — resume cap bounds per-DIR not per-CAUSE — CLOSED WITH REGRESSION
Found round 2 (L6, superseding R1-29). Round-3 repair: the wrapper normalizes each death's
abort reason into a dead SIGNATURE (exit class + first abort-record line); after M=3
consecutive fresh-dir deaths with the same signature it writes a HALT marker (preflight refuses
to launch until the human clears it); the "cannot ... forever" claim softened to per-dir scope
(report §3.4, line 409). Verified round 3 by direct read. Regressions: (1) abort records are
dated + run-dir-addressed by design, so unless the (asserted, unspecified) normalization strips
that variable content, no two signatures match and the HALT never fires — identity-keyed-
detector-lineage-blind applied to the design's own repair (L5 -> R3-11); (2) the per-night burn
arithmetic "k×$5/night" does not follow from one-resume-per-fire semantics (L2 -> R3-12). Closure
class: **closed_with_regression** -> successors **R3-11, R3-12**.

### R2-11 — 30-day auto-stale composes badly with graduation latency — CLOSED WITH REGRESSION
Found round 2 (L6, superseding R1-22). Round-3 repair: a human-set `graduation-queued` status
exempts a stub from the 30-day auto-stale while still deduping its class (report line 267).
Verified round 3 by direct read. Regression: the exemption has NO upper bound — an abandoned
queued stub (queue-and-forget, operator turnover) is exempt forever while still deduping, so it
permanently subtracts its signal class with no backstop, re-opening for the queued subset the
exact monotonic-blinding failure R1-22 was minted to prevent (L6 -> R3-13). Closure class:
**closed_with_regression** -> successor **R3-13**.

### R2-12 — idempotency/resume gate still read loop-writable state — CLOSED
Found round 2 (L5, superseding R1-19). Round-3 repair: the wrapper derives completeness from
its own step-7 record in the operator-owned dir (not the loop-writable run dir), with the DEAD
marker location stated. Verified round 3 by L6 (sound). Closure class: **closed**.

### R2-13 — §6 row 5 likelihood cell carried the stale "(no API)" — CLOSED
Found round 2 (L3, superseding R1-9). Round-3 repair: cell requalified "(no spend-limit API;
rate-limit API unreachable at this auth tier — §5.1/R1-9, cell requalified round 2, R2-13)"
(report line 1122). Verified round 3 by direct read. Closure class: **closed**.

### R2-14 — §7 self-audit still characterized pricing as MEDIUM — CLOSED
Found round 2 (L4, superseding R1-11). Round-3 repair: the §7 Pattern B/E bullet now carries
"(upgraded to leaf-verified HIGH round 1, R1-11 — this bullet's lag fixed round 2, R2-14)" and
R1-11 is added to §7's banked-upgrade list. Verified round 3 by L4 (present, no residual).
Closure class: **closed**.

### R2-15 — §3.4 rung-1 label still "RECOMMENDED default" unqualified — CLOSED
Found round 2 (L5, superseding R1-14). Round-3 repair: label qualified "RECOMMENDED default
AMONG SCHEDULED RUNGS, once the human opts in." Verified round 3 by L5 spot-check. Closure
class: **closed**.

### R2-16 — gate-survival table rung-0 cells overstated + manual spend out-of-ledger unstated — CLOSED WITH REGRESSION
Found round 2 (L5, superseding R1-27). Round-3 repair: the R0 L2 cell split ("fence YES (cache
copy); canary n/a — no wrapper at rung 0", report line 774) and manual-run spend declared
out-of-ledger by design with the cap-arithmetic composition stated. Verified round 3 by direct
read. Regression: the rung-0 EXECUTION SHAPE for steps 0/2/4/7 (all wrapper-hosted) is still
undefined while §1.4 calls rung 0 "same code path" — the unowned sibling of the ledger horn
(L5 -> R3-2, jointly with R2-5/R2-6). Closure class: **closed_with_regression** -> successor
**R3-2**.

### R2-17 — §3.3 granted and revoked pdf/arxiv ToolSearch reach in one sentence — CLOSED
Found round 2 (L5). Round-3 repair: blue picked the trade — the loop's MCP profile is
`--strict-mcp-config --mcp-config <sleeper-mcp.json>` naming qmd only, and the round-1
parenthetical "research subagents reach pdf/arxiv tools via ToolSearch" is corrected as false
under that flag (ToolSearch discovers only from DECLARED servers; strict-mcp-config ignores the
project `.mcp.json`) — report lines 676-680. Verified round 3 by direct read. Closure class:
**closed**.

### R2-18 — $2–5/night × 30 vs ~$50 cap unreconciled — CLOSED
Found round 2 (L5). Round-3 repair: expected per-run spend owned from the smoke figure, and the
ceiling-vs-cap composition stated as the intended month-end throttle / anomaly signal (not
death). Verified round 3: L1 recomputed (30×$0.10–0.50 = $3–15/mo, ≥3× cap headroom at the
band; $2–5×30 = $60–150 ceiling correctly owned as throttle) and L5 concurs. Closure class:
**closed**.

### R2-19 — §0 artifact enumeration omitted the skill file + plugin manifest — CLOSED WITH REGRESSION
Found round 2 (L1). Round-3 repair: enumeration made total over the round-1 tree (skill file +
manifest added; L1 recount = 8 entries, total, HIGH). Regression: the round-2 minted artifacts
(SessionStart staleness hook + hooks.json, the doctor-line delta EDITING prosthetic-conscience,
the two operator-owned JSON configs) fall outside the now-"total" enumeration — exhaustive-
sweep-omits-own-specimen recurring against the round-2 additions (L5 -> R3-10). Closure class:
**closed_with_regression** -> successor **R3-10**.

### R2-20 — [^Backlog] pin range "15–17" under-covered by one line — CLOSED
Found round 2 (L1). Round-3 repair: range -> "15–18". Verified round 3 by L5 spot-check. Closure
class: **closed**.

### R2-21 — "The DGM analogy is exact" overstated — CLOSED
Found round 2 (L2). Round-3 repair: "exact"->"direct" plus the honesty clause (DGM evaluates
every change before archiving but admits even low scorers for exploration; our promotion gate
is stricter, pass-required). Verified round 3 (L2) by live re-fetch of arxiv.org/html/2505.22954:
"Only agents that compile successfully and retain the ability to edit ... are added to the
archive"; low scorers deliberately retained; parent selection roughly proportional to score —
every leg corroborates, HIGH. Closure class: **closed**.

### R2-22 — one [^CostRecord] marker on two figures from different artifacts — CLOSED
Found round 2 (L3). Round-3 repair: [^EfficiencyPlan] added beside the run-3 $149.95 figure
(cost.md carries $414.97). Verified round 3 by L5 spot-check. Closure class: **closed**.


## Round 4 closures (red merge, 2026-07-17)

### R3-1 — canary --json-schema leg undocumented + unprobed behavioral legs — CLOSED WITH REGRESSION
Found round 3 (L2+L5, superseding R2-1). Round-3 revision: blue demoted the mid-drive
`--json-schema` leg to verify-at-build with a fenced-block text-parse fallback named as the design
of record, and added OQ22 (mid-drive structured output) + OQ23 (two-phase-drive behavioral
acceptance legs). Verified round 4 (L2): step-3 restatement faithful, no regression on the
structured-output leg — HIGH. Regression: the R3-1 degrade-note's "named readers" are unspecified
at both reader sites — §2.3's `confidence` field names only the R2-17 PDF caveat (no qmd-degrade
labeling obligation), and §3.4's doctor line prints skip/abort reasons (a qmd degrade is neither),
so the reader declaration exists only at the writer's site (L2 → R4-7). Closure class:
**closed_with_regression** -> successor **R4-7**.

### R3-2 — rung-0 execution shape undefined ("same code path" vs wrapper-less) — CLOSED WITH REGRESSION
Found round 3 (L5, superseding R2-5/R2-6/R2-16). Round-3 revision: blue adopted the lead-endorsed
R2-16b resolution — rung 0 invokes the SAME wrapper via `node sleeper-wrapper.mjs --manual`, so
"same code path" becomes true by construction, the out-of-ledger accept DISSOLVES, and the rung-0
marker/corroboration voids close in the same move (debate.md ### BLUE round 3). Verified round 4
(L1 cross-read §3.4 ladder row 0 + §2.2 step 0 — CONSISTENT; manual dirs marker-stamped).
Regression (a DISTINCT surface, not the manual-execution path): `/self-improve` ships as an
invocable command under `commands/` carrying the full loop payload, guarded only by a prose header,
with no `disable-model-invocation: true` (its sibling graduate.md carries it), so the unwrapped path
stays mechanically open via command invocation (human or model/scheduled); and §3.4's containment
sentence has INVERTED polarity — markerless out-of-contract dirs PROVIDE non-sleeper corroboration
rather than "requiring" it (L5+L6 → R4-1). Origin-tag escape's 4th surface. Closure class:
**closed_with_regression** -> successor **R4-1**.

### R3-3 — origin cap does not reach the red-memory surface — CLOSED WITH REGRESSION
Found round 3 (L6, superseding R2-5). Round-3 revision: blue extended provenance tagging to the
red-memory mirror — patterns appearing/changing in a wrapper-logged sleeper window are appended to
a `sleeper-authored-patterns` list, capped at 1 per class, excluded from the corroboration pool;
snapshots extended to the red-memory dir; mirror pre-run frozen. Verified round 4 (L1 cross-read
§2.2 step 0 + §6 row 10 — CONSISTENT with harvest timing). Regression, two legs: (a) FORECLOSED
PREMISE — the nightly red-merge seat CANNOT write `.claude/agent-memory/**` under the sleeper
profile (fence outside research/+ideas/, `Edit(<REPO>/.claude/**)` deny, phase-0 canary abort on
marker loss), so §1.5's premise (the seat writes the memory dir in-window) and §4.2's profile
cannot both be true; the real consequence (the seat's mandatory record-new-patterns write
fails-denied every night — a silent degradation of the nightly adversary's learning + a routine
fence-denial stream in the fired-record) is stated nowhere; R3-3's live remit shrinks to
out-of-window authorship and the snapshot-diff machinery is dead where aimed. (b) CHANGE-KEYED
GRANULARITY — the trigger is file/header APPEARS-OR-CHANGES applied to the whole pattern, but red's
memory discipline is to EXTEND existing files, so one nightly append converts a pattern carrying
months of human-present evidence to sleeper-origin (R1-22 monotonic-blinding through the guard).
(L5+L6 → R4-4.) Closure class: **closed_with_regression** -> successor **R4-4**.

### R3-4 — run-window fallback void for dead runs — CLOSED WITH REGRESSION
Found round 3 (L1, superseding R2-5). Round-3 revision: blue logs the run-window START in the
operator-owned dir at step 0 and the END at whatever exit it observes (step-7 record, abort, or
DEAD-mark time), and runs the origin-marker porcelain sweep at DEAD-marking time — so even a
resumeless death leaves its window on record. Verified round 4 (L1 both required-fix legs present).
Regression the repair itself mints: a wrapper hard-kill (SIGKILL/power loss/OS kill — §6 row 6's
own High-likelihood environment) is an UNOBSERVED exit, leaving a window with START and no END
until DEAD-mark time ~3–4 days later; the multi-day window then sweeps daytime HUMAN-present run
dirs into `origin: sleeper` (capped at 1, excluded from the corroboration pool — suppressing
exactly the human-present corroboration ordinary classes wait on), and harvest's reading of an
UNCLOSED window is unspecified (L1+L6 → R4-6). Closure class: **closed_with_regression** ->
successor **R4-6**.

### R3-5 — infrastructure-class classifier source unstated — CLOSED
Found round 3 (L6, superseding R2-6). Round-3 revision: blue stated that the infrastructure-class
tag is assigned SOLELY from the wrapper's own event log (mechanical, session-unforgeable); friction
TEXT never self-classifies into the no-corroboration lane. Verified round 4 (L1): verbatim sentence
present, CLEAN. Closure class: **closed**.

### R3-6 — doctor line "also surfaces independently" overclaimed (last-slot only) — CLOSED
Found round 3 (L1+L5, superseding R2-6). Round-3 revision: the doctor/dead-man line now prints
PER-SIGNATURE COUNTS since last human clear, keyed by the R3-11 normalization (one mechanism fixes
R3-6 + R3-11 as the lead directed), so a transient infra event followed by another is no longer on
the stub flag alone. Verified round 4 (L1 cross-read §3.4 dead-man paragraph — CONSISTENT). Closure
class: **closed**.

### R3-7 — abort-path watchman gap (no END snapshot after a dead run) — CLOSED
Found round 3 (L5, superseding R2-7). Round-3 revision: the wrapper snapshots at EVERY exit path it
observes (abort included — it outlives the session by construction), and the cross-run compare is
defined against the LAST RECORDED snapshot of any kind, failing closed on a missing chain link per
invariant 8. Verified round 4 (L5 spot-check): a crash-without-end-snapshot is covered by
compare-vs-last-recorded = the dead run's START snapshot, spanning the window — sound. (The
unobserved-hard-kill window-log residual rides on R3-4's successor R4-6, not here — the snapshot
chain leg holds.) Closure class: **closed**.

### R3-8 — leg (b) detection scoped to in-repo porcelain + guardrail hashes — CLOSED
Found round 3 (L6, superseding R2-7). Round-3 revision: blue scoped leg (b) honestly — the snapshot
detects in-repo working-tree deltas and guardrail-file tampering, NOT arbitrary out-of-repo
subprocess writes; the out-of-repo write is stated as a residual bounded by pinned code +
no-remaining-write-gadget (citable only post-R3-15), sequenced as the lead directed. Verified round
4: the body carries the edits at layer 4 (iv) ("bounded HONESTLY round 3 (R3-8)"), layer 5
("Sensing scope stated honestly"), and §6 row 4 leg (b). The CHANGELOG Round-3 enumeration omits
the R3-8 bullet (16 of 17 under an "all 17 addressed" header) — a navigation-hint defect, not a
report-body defect; raised separately as R4-16 (change-summary desync, 2nd consecutive round).
Closure class: **closed**.

### R3-9 — wrapper absent from protection enumerations; root of trust unverified — CLOSED WITH REGRESSION
Found round 3 (L5+L6, recommend-not-block). Round-3 revision: blue absorbed the cheap parts —
watchmen named in layer 3 + the snapshot hash set, absolute-path invocation from the operator-owned
copy, a SessionStart last-check timestamp, wrapper-hash dates on the doctor line — and OWNED the
self-verification limit in text (the wrapper cannot attest itself; the residual is a box-compromise
scenario, not a loop-autonomy one). Verified round 4 by direct read of the round-3 debate/report
text. Regression: the dead-man push channel's disarm switch (scheduling-enabled flag) has unstated
custody, and the same-round R3-9 (always-print liveness) and R3-10 (condition-on-flag) repairs to
the SAME SessionStart line have unstated composition — if "conditioned" means fully silent when the
flag is off/absent, flag loss silently disarms the dead-man channel and a disarmed hook looks
identical to a healthy-but-disabled one, re-opening what R3-9's always-print clause closed (L6 →
R4-14). Closure class: **closed_with_regression** -> successor **R4-14**.

### R3-10 — round-2 minted artifacts outside §0's "total" enumeration — CLOSED WITH REGRESSION
Found round 3 (L5, superseding R2-19). Round-3 revision: blue extended §0's enumeration (SessionStart
staleness hook + hooks.json, the CROSS-PLUGIN prosthetic-conscience doctor-line delta, the two
operator-owned config files) and conditioned the SessionStart warning on scheduling-enabled.
Verified round 4 (L1 items present). Regression: the R3-10 repair extended the enumeration but never
reconciled the count HEADLINE — the same §0 paragraph simultaneously asserts "exactly THREE new code
artifacts" and enumerates a fourth executable (R3-10's own text: "a new executable + hooks.json
registration"); the skill file/manifest got an explicit "new PROSE artifacts, not code"
classification, the SessionStart hook got none, and its host plugin is unstated (L1 → R4-8).
Closure class: **closed_with_regression** -> successor **R4-8**.

### R3-11 — dead SIGNATURE normalization unspecified (HALT never fires) — CLOSED
Found round 3 (L5, superseding R2-10). Round-3 revision: blue specified the normalization —
signature = exit class + templated first abort line with `<date>`/`<path>`/`<id>`/`<n>` placeholders
(dates/paths/ids/nonces/digits stripped, the corpus's identity-keyed-detector lesson applied to the
design's own repair) — added zero-HALT-firings telemetry to the doctor line so a never-firing
detector is visible, and owned the alternating-cause (A,B,A,B) residual. Verified round 4 (L1/L2/L5
concur: concrete spec present, telemetry present, residual owned). Closure class: **closed**.

### R3-12 — per-night burn arithmetic wrong (k×$5 vs one-resume-per-fire) — CLOSED WITH REGRESSION
Found round 3 (L2, superseding R2-10). Round-3 revision: blue stated the in-night retry semantics
and recomputed (≤$5/night; cap trips ~10 nights; HALT at night 12 or the cap, whichever first).
Verified round 4 (L2 recomputed: one resume per fire → DEAD night 4/dir → HALT night 12; cap ~night
10 — the printed per-night figure now correct). Regression (merge recompute, resolving an L2/L6-held
vs L5-disputed conflict): the recomputed bound treats cap-trip as terminal, but the monthly cap
RESETS — at ceiling pricing the cap trips FIRST (~night 10) and PAUSES death accrual, so the third
same-signature death cannot occur in month 1 and the HALT lands early the NEXT month; worst-case ≈
$55–60 across two months and there is no "whichever comes first" race (the cap always pre-empts,
then un-pre-empts at rollover). The bounded conclusion survives (≈ one cap + ε); the printed race
and single-month bound are wrong — the same repair-introduced-arithmetic class R3-12 itself fixed
(L5 → R4-10). Closure class: **closed_with_regression** -> successor **R4-10**.

### R3-13 — graduation-queued exemption has no upper bound — CLOSED WITH REGRESSION
Found round 3 (L6, superseding R2-11). Round-3 revision: blue gave `graduation-queued` its own
M=90-day `queued-stale` re-surface for human re-confirmation (§1.4/§2.3), and labeled the field
"no status is timer-free." Verified round 4 (L1: both sites present, CLEAN for the queued state).
Regression: "no status is timer-free" is false as written — the terminal states `rejected` and
`graduated` have neither timer nor stated dedupe semantics, so a rejection either permanently
subtracts its class (the monotonic-blinding failure R1-22 forbids) or re-mints a stub every run
(the Dependabot-fatigue arm); the backlog regression rule covers backlog items, not stub statuses.
Third consecutive per-status patch (R1-22 → R2-11 → R3-13) signals the missing root invariant
(every status's dedupe effect has a stated re-surface path) (L5 → R4-9). Closure class:
**closed_with_regression** -> successor **R4-9**.

### R3-14 — dontAsk closed-world premise refuted; Bash read carve-out — CLOSED WITH REGRESSION
Found round 3 (L2, first-raise). Round-3 revision: blue corrected §3.2's dontAsk sentence to carry
the carve-out, re-derived §4.2's profile with enumerated `deny` rules over the 14-command read-only
set (all 14 deny-covered, diffed line-by-line by L4/L6), re-argued §6 rows 4/13 + layer 2 on the
true surface, and added [^PermissionsDoc] round-3 quotes (verbatim vs live doc, L2/L3/L4/L6 HIGH,
zero drift). Verified round 4. Regressions, two: (a) the enumeration rests on a NON-EXHAUSTIVE doc
list — the doc says the set "**include**[s]" 14 commands and never claims completeness, and the same
page names `sort`/`sed` as classifier-reasoned commands not in the 14; an unlisted member
(`sort`/`file`/`readlink`/`strings`/`less`) auto-runs un-denied, re-opening the Bash read channel
(invariant-soundness-by-enumeration applied to the design's own repair) (L3 → R4-3); (b) the
prior-exposure example is refuted at the leaf — `Bash(cat //…/.claude/projects/…)` on row 13's named
transcript target was NOT auto-approved under the round-2 profile, because that profile's
`Read(//…/.claude/projects/**)` deny extends to recognized Bash file commands per the doc's own
deny-reach clause; the real round-2 exposure was allow-scoped-but-not-Read-denied paths
(credentials-class), so both sites overstate what R3-14 closed (postmortem-misdiagnosis) (L2+L4 →
R4-5). Closure class: **closed_with_regression** -> successors **R4-3, R4-5**.

### R3-15 — git log --output write gadget refutes "pinned read-only git argv" — CLOSED WITH REGRESSION
Found round 3 (L4, first-raise; tool-run confirmed). Round-3 revision: blue added
`--output`/`--output-directory`/`-O`-class flags to the sleeper-guard hook's Bash-write matcher and
belt denies, added the channel to OQ18's named scope, and downgraded row 4 leg (a) + layer 4 (i) to
"read-only EXCEPT the write channel." Verified round 4. Regressions, two: (a) read-only git is
retained UN-enumerated in the carve-out (the design needs git reads), so "deny-enumerated per
command" overclaims for its git member — and sibling git-native output flags escape both belt and
hook: leaf-verified this box `git format-patch -1 -o /tmp/l5probe` → exit 0, arbitrary out-of-repo
patch (the `-o` short form matches none of the three long-form belt denies and is not in the hook's
named list); `git archive -o`/`git bundle create` are further writers — sibling-halo on the R3-15
closure, out-of-repo targets in the R3-8 blind spot (L5 → R4-2); (b) the "no prompt" reproduction is
attributed to the carve-out classifier without isolating the layer — both round-3 reproductions ran
under `defaultMode: "auto"`, where the AUTO classifier is the approving layer, so "showing the
carve-out classifier itself passes `--output`" and "rule-pinning alone cannot close it" are not
established by the probe (the isolating dontAsk-zero-allow probe was attempted twice from lens seats
and DENIED by the seat's own auto-mode classifier; not routed around) (L3+L4 → R4-11). Closure class:
**closed_with_regression** -> successors **R4-2, R4-11**.

### R3-16 — §1.3 telemetry-input row stale (0.6.0 future tense) — CLOSED
Found round 3 (L1, first-raise). Round-3 revision: blue updated the row to "SHIPPED as of FEOV 0.7.0
— present in this run's own trajectories/." Verified round 4 (L1): Glob confirms
`trajectories/board-telemetry.jsonl` EXISTS in this run dir — filesystem-verified, HIGH. Closure
class: **closed**.

### R3-17 — [^Pricing] tokenizer scope under-inclusive (omits Opus 4.7+) — CLOSED
Found round 3 (L3, first-raise; recommend risk-accept). Round-3 revision: rather than risk-accept,
blue completed the set — "tokenizer set completed with Opus 4.7+" (debate.md ### BLUE round 3).
Verified round 4 (L3 live re-fetch): the pricing page names "Opus 4.7 and later Opus models, Fable
5, Mythos 5, Mythos Preview, Sonnet 5" — verbatim, HIGH, zero drift. Closure class: **closed**.

## Round 5 closures (red merge, 2026-07-17)

### R4-1 — /self-improve invocable-command surface + inverted containment polarity — CLOSED WITH REGRESSION
Found round 4 (L5+L6, superseding R3-2). Round-4 revision: `self-improve.md` became a thin
trampoline carrying `disable-model-invocation: true` (matching graduate.md), the loop payload
moved to the wrapper's phase-1 prompt sourced from skills/continuous-learning, and §3.4's
containment polarity was corrected (markerless out-of-contract dirs are NON-sleeper and CAN
corroborate; residual = a human's deliberate paste-run, named, quarantine deferred as OQ24).
Verified round 5: L6 audited the trampoline mechanism directly — sound, symmetric with
graduate.md; L5/L2 verified the §0 tree + §3.4 paragraph + §6 row 10 + OQ24. Regressions:
(a) two body sites still specify the OLD payload shape — §3.4 ladder row 0 ("the /self-improve
command markdown is the wrapper's phase-1 prompt payload in EVERY mode") and §3.3's adopted
Phase-4 acceptance test (`claude -p "/self-improve"` produces a run dir — now fails by
construction; cheapest pass re-inlines the payload) (L2+L5 → R5-2); (b) the paste-run pool
residual was deferred via OQ24 on a rationale that prices only paste-run likelihood — the
rung-2 automatic-poisoning case is unargued, folded into the pool root-invariant finding
(L6 → R5-3). Closure class: **closed_with_regression** -> successors **R5-2, R5-3**.

### R4-2 — git carve-out member un-enumerated; sibling git-native writers escape belt+hook — CLOSED WITH REGRESSION
Found round 4 (L5, superseding R3-15). Round-4 revision: the git channel INVERTED to a
read-ALLOWLIST at the sleeper-guard hook (deny any git argv not an exact allowed read form —
the lead-endorsed invariant-6 shape); belt denies extended to `-o`/`-O` short forms and the
format-patch/archive/bundle/config/gc/repack/maintenance writer family; §6 row 4 corrected to
name the git exception with OQ18(c) as standing test. Verified round 5: L3 reproduced the
spaced-form gadget live (`git format-patch -1 -o /tmp/… → exit 0` out-of-repo) confirming the
absorbed leaf claim; L5/L6 verified the allowlist inversion text. Regressions: (a) R4-2's own
reachability premise — "Where Bash IS reachable (a rebuilt rung, the Workflow seat agents,
profile drift)" — contradicts §4.3 layer 4 (iii)'s "seat agents are full permission-engine +
hook subjects"; the seat-surface composition with R4-3's bare deny is undecided and
settings-inheritance unprobed (L3+L5+L6 → R5-1); (b) the ATTACHED `-o<value>` form (no space)
escapes the new `Bash(* -o *)` belt pattern — leaf-verified by L6 and re-reproduced at the
merge seat (`git format-patch -1 -o/tmp/r5mergeA HEAD` → exit 0, out-of-repo patch) — the
enumerate-and-extend regress recurring inside the repair; belt-only (hook allowlist is the
load of record) but the belt binds rebuilt rungs 3–4 (L6 → R5-5). Closure class:
**closed_with_regression** -> successors **R5-1, R5-5**.

### R4-3 — carve-out deny enumeration rests on a non-exhaustive doc list — CLOSED WITH REGRESSION
Found round 4 (L3, superseding R3-14). Round-4 revision: a bare `Bash` deny added to the
sleeper profile — doc-verified ("A bare tool name like `Bash` removes the tool from Claude's
context entirely"; `Bash(*)`≡`Bash` as deny) — closing the whole dontAsk read-only carve-out
class (enumerated + unlisted members + read-only git + every git write gadget) at the tool
boundary for the top-level session; enumerated denies retained as belt; `sort`/`sed`/`file`/
`readlink`/`strings`/`less` + `~/.ssh`/`.env` belt denies added; OQ18(c) gained the
member-enumeration probe. Verified round 5: L3 re-fetched the permissions doc live — every
load-bearing quote verbatim, zero drift; the "include[s]"/non-exhaustive premise and
deny-supremacy confirmed. Regression: the structural close is scoped to the TOP-LEVEL session
("§2.2's session steps never invoke Bash") while §6 row 13 and §4.3 layer 2 state TOTAL
claims ("holds on the Bash channel too") and §4.2's R4-2 bullet simultaneously concedes Bash
IS reachable in the Workflow seat agents — the two horns are undecided, and if seats do not
inherit the `--settings` profile, closure for the nightly worker population reverts to the
enumeration R4-3 itself declared non-load-bearing (L3+L5+L6 → R5-1; the stale "carve-out
auto-approves read-only git regardless" comment is the same finding's trivial sub-leg).
Closure class: **closed_with_regression** -> successor **R5-1**.

### R4-4 — provenance mechanism aimed at a profile-foreclosed write — CLOSED WITH REGRESSION
Found round 4 (L5+L6, superseding R3-3). Round-4 revision: blue picked the cheapest coherent
horn and stated it plainly — agent-memory writes are DENIED by design under the sleeper
profile; nightly seats do not learn; nightly patterns reach the corpus only via
ordinary-classed friction text; the R3-3 machinery re-scoped to belt-for-drift; tagging at
window-ADDED granularity (pre-existing content keeps its origin), closing the R1-22
monotonic-blinding leg. Verified round 5 (L5/L6 direct read of the rewritten §1.5 paragraph +
§6 row 10). Regression the repair mints: the mechanism making the denial stream un-alarming —
bucket every `.claude/agent-memory/**` denial as `origin=nightly-seat,
class=expected-memory-deny`, unflagged by the doctor line — is keyed on the TARGET PATH, not
the actor/cause, so an injection-driven memory-poisoning ATTEMPT (§6 row 14's
Medium-likelihood channel) at any volume is indistinguishable from the seat's one expected
protocol write; enforcement holds, attack-attempt telemetry on the adversary-steering surface
is erased (L5+L6 → R5-6). Closure class: **closed_with_regression** -> successor **R5-6**.

### R4-5 — prior-exposure example refuted by the doc's deny-reach clause — CLOSED
Found round 4 (L2+L4, superseding R3-14). Round-4 revision: postmortem corrected at BOTH
sites (§4.2 prose bullet AND §6 row 13) — the round-3 `Bash(cat …/.claude/projects/…)`
example acknowledged as mis-diagnosed (that path was Read-denied, and Read/Edit denies extend
to recognized Bash file commands per the doc); example re-pointed at un-named
credentials-class paths (`~/.aws/credentials`, stray `.env`); deny-reach clause noted in the
profile comments. Verified round 5 (L3): the deny-reach clause re-fetched verbatim live
("Read and Edit deny rules apply … and to file commands Claude Code recognizes in Bash, such
as `cat`, `head`, `tail`, and `sed`"); both corrected sites faithful to the doc. Closure
class: **closed**.

### R4-6 — unobserved-exit window sweeps human-present dirs — CLOSED WITH REGRESSION
Found round 4 (L1+L6, superseding R3-4). Round-4 revision: a window's END is additionally
bounded by the NEXT wrapper START (no window spans invocations); a DEAD-mark-closed window is
flagged `retroactive-uncertain` with its markerless sweep CONFINED to sleeper date-key naming
(other dirs surfaced for human confirmation); unclosed-window read specified
(extends-to-present for date-key-named dirs only); snapshot-chain + resume backstop owned in
step-7/layer-5 text. Verified round 5 (L5 direct read; all three owed legs present).
Regressions: (a) the confinement decides sweep membership BY dir name while §1.5's unqualified
doctrine sentence ("reads ONLY the marker, never the dir name") still stands, and the named
convention "the wrapper's own sub-run slug" is unknowable after a hard-kill (model-chosen
nightly, recorded nowhere durable; mkdir-to-stamp bound unstated) — row 10's "cannot sweep
human-present dirs" overclaims or the auto-tag under-delivers (L5 → R5-4); (b) the
window-close is the second of three per-surface pool patches folded into the pool
root-invariant finding (L6 → R5-3). Closure class: **closed_with_regression** -> successors
**R5-4, R5-3**.

### R4-7 — degrade-note reader sites carried no surfacing obligation — CLOSED
Found round 4 (L2, superseding R3-1). Round-4 revision: §2.3's confidence field gained the
"recall-degraded (qmd down; lexical Grep/Read only this run)" labeling clause; §3.4's doctor
line gained the qmd-degrade streak term ("qmd degraded on M of the last N runs"); §2.2 step
0's "named readers" sentence now has both readers specified. Verified round 5 (L2 + L5 direct
reads at both reader sites — the obligation exists where readers read). Closure class:
**closed**.

### R4-8 — "exactly THREE code artifacts" unreconciled with the fourth executable — CLOSED WITH REGRESSION
Found round 4 (L1, superseding R3-10). Round-4 revision: count corrected to FOUR code
artifacts (SessionStart staleness-warning hook counted); host plugin named — sleeper-service,
with the rationale (fires in interactive sessions independent of the scheduler, so it ships
with the sleeper plugin's own hooks.json). Verified round 5 (L5/L6 direct read of §0).
Regressions: (a) the §0 TREE (labeled "the implementable shape") was never reconciled — no
hooks/hooks.json entry, no SessionStart executable, and nothing registers the sleeper-guard
PreToolUse hook either; the enumeration has now been repaired three times (R2-19/R3-10/R4-8)
while the tree never was (L5 → R5-8); (b) homing the hook in sleeper-service's OWN hooks.json
extends the empty-bin crash-storm surface to every interactive session while §6 row 9's bound
cites prosthetic-conscience's hooks.json — guard coverage for the new file unstated
(L6 → R5-7). Closure class: **closed_with_regression** -> successors **R5-8, R5-7**.

### R4-9 — rejected/graduated statuses timer-free; missing root invariant — CLOSED WITH REGRESSION
Found round 4 (L5, superseding R3-13; chain root R1-22). Round-4 revision: the root invariant
stated ONCE in §1.4 — every status's dedupe effect has a stated re-surface path; no status
subtracts its class permanently — with terminal-state semantics: `graduated` → class
recurrence re-enters flagged `regression`; `rejected` → dedupes for a 90-day/rate window then
re-surfaces `rejected-recurring` for one-keystroke re-confirm. §2.3 enum + §6 row 3 carry it.
Verified round 5 (L5: invariant present once, semantics stated; L2: enum cross-read).
Regressions: (a) both arms of the `rejected` clause key on a rejection DATE no artifact
records (stub filenames date the MINT; status edits undated) — "pre-rejection rate" and the
90-day window are uncomputable by the zero-token harvest as specified (L5 → R5-9); (b) the
`regression` token is not in the §2.3 enum and its domain (docket flag vs stub status; the
`rejected-recurring` setter) is unpinned (L2 → R5-9). Closure class:
**closed_with_regression** -> successor **R5-9**.

### R4-10 — cap-trip treated as terminal but the monthly cap RESETS — CLOSED
Found round 4 (L5, superseding R3-12; intra-round lens conflict resolved for L5 by merge
recompute). Round-4 revision: §3.4 recomputed — at ceiling pricing deaths land nights 4 and 8;
the ledger preflight cap-skips from ~night 11; no death accrues during cap-skips, so the M=3
HALT lands early the NEXT month (~$5–10 in); worst-case ≈ one monthly cap + ≤2 nights
(~$55–60 across two months); the "whichever comes first" race removed. Verified round 5 by
THREE independent recomputes (L2, L5, L6) — all figures check. Closure class: **closed**.

### R4-11 — "no prompt" attributed to the carve-out classifier without layer isolation — CLOSED
Found round 4 (L3+L4, superseding R3-15). Round-4 revision: attribution re-scoped at all
three sites (§4.2 bullet, §7 round-3 update, OQ23(d)) to "consistent with carve-out
classification but does NOT isolate it (both probes ran under defaultMode: auto)"; the
isolating dontAsk-zero-allow probe deferred to build (OQ18(c)/OQ23(d)); hook matcher stated
chosen-conservative. Verified round 5 (L5 direct read of all three sites — faithful). Closure
class: **closed**.

### R4-12 — est_complexity factor had no stated input source — CLOSED WITH REGRESSION
Found round 4 (L5, first-raise). Round-4 revision: §1.4 stage 2 now states the source —
default 1 (inert; the factor vanishes) unless the class's matching ideas/backlog.md entry
carries a human-recorded complexity note, which harvest parses; ranking stays zero-token
arithmetic. Verified round 5 (L5: clause present and safe-defaulted). Regression: "harvest
parses that note's value" states no FORMAT (backlog entries are free prose — unparseable as
arithmetic), and L1's leaf check at the pin shows NO backlog entry carries any parseable
complexity field — the named source is a convention that does not yet exist, so the divisor
is universally inert while the text implies it activates (L1+L5 → R5-10). Closure class:
**closed_with_regression** -> successor **R5-10**.

### R4-13 — provenance/corroboration gate absent from the gate-survival table — CLOSED WITH REGRESSION
Found round 4 (L5, first-raise). Round-4 revision: the §3.4 gate-survival table gained a
"Provenance/origin-tag + corroboration gate (§1.5)" row (YES R0/R1 via the wrapper; NO R2–R4
absent rebuild) and the graduation-grade adoption requirement names rebuilding this gate as a
precondition. Verified round 5 (L5: row present, requirement named). Regression: the row +
instructional adoption requirement is the third per-surface patch on the negatively-defined
corroboration pool (R2-6 → R3-3/R3-5 → R4-1/R4-6/R4-13) with the root invariant still
missing; at rung 2 the poisoning is AUTOMATIC (local runs against the same harvested corpus,
markerless → counted non-sleeper next morning) behind an instructional-only gate — the class
§4.1's own 0/175 evidence refutes; OQ24's deferral rationale never argues this case
(L6 → R5-3). Closure class: **closed_with_regression** -> successor **R5-3**.

### R4-14 — dead-man disarm flag custody + R3-9/R3-10 composition unstated — CLOSED WITH REGRESSION
Found round 4 (L6, superseding R3-9). Round-4 revision: the scheduling-enabled flag lives in
the OPERATOR-OWNED dir beside the ledger (outside every session write surface); the
SessionStart line is NEVER fully silent — flag-off prints the disabled token, flag-absent
prints "sleeper: flag missing" (anomaly), flag-on-stale prints the staleness warning —
silent-disarm distinguishable from disabled-by-choice. Verified round 5 (L5: four states all
print; custody stated; composition with R3-9 explicit). Regression: flag-ABSENT is the
DEFAULT state of a fresh install before schedule-setup ever runs, so the anomaly print nags
every never-configured operator per session — never-configured vs flag-lost undistinguished;
the alert-fatigue mode the report catalogs, aimed at its own dead-man channel (L6 → R5-7).
Closure class: **closed_with_regression** -> successor **R5-7**.

### R4-15 — OQ18(c) flag-scoped while the black box is the git SUBCOMMAND boundary — CLOSED
Found round 4 (L4, first-raise; recommend-not-block). Round-4 revision: OQ18(c) extended to
the subcommand boundary of "read-only forms of git" with `config`/`gc`/`repack`/`maintenance`
named as probe cases (the `git config` pager/alias arbitrary-command-exec consequence
stated); the hook git read-ALLOWLIST rejects them by default (inside deny-by-default), and
the writer family entered the belt denies. Verified round 5 (L5/L3: OQ18(c) text + §4.2
carry it). Closure class: **closed**.

### R4-16 — CHANGELOG round-3 enumeration missing the R3-8 bullet — CLOSED
Found round 4 (L5, first-raise; process note). Round-4 revision: the R3-8 bullet added to
CHANGELOG Round 3 (content copied from §7's round-3 list). Verified round 5 (L5 + merge
direct read of CHANGELOG — bullet present; the change-summary channel matches the report for
the first time in three rounds). Closure class: **closed**.

---

## Debate record

Literal transcript: `debate.md` (this run dir). Per-round synopsis:

- **Round 0 — BLUE synthesis.** Union of three method-lens lanes (adversarial-disconfirming, primary-literature, local-probe). All five frontier hypotheses SUPPORTED, three with carried conditions; 132 claims, 62 minority-tagged, one cross-lane conflict carried openly; 15 open questions. The R1-7 pinned-path defect declared as standing friction from the start.
- **Round 1 — RED FAIL: 30 gaps, mass 118.75.** The evidence base graded "unusually sound" (~100 citation pairs leaf-verified); the failure mode is composition, not citation — write-capable subprocess channel (R1-16), self-authored harvest inputs (R1-25), gate stack absent at rungs 2–4 (R1-27), fence fails open on marker loss (R1-28). **BLUE:** all 30 addressed additively; harvest moved wrapper-side; ledger moved operator-owned; nightly profile hardened.
- **Round 2 — RED FAIL: 23 gaps, mass 81.5.** The repairs landed; the repair MACHINERY is the board (canary certifies the wrong thing R2-2; same-round fixes don't compose R2-3, the run's one HIGH; evidence with no watchman R2-7/8/9). **LEAD (first docket):** R1-7 risk_accepted (run-scope, fix routed to run infrastructure); R2-1..R2-16 carried with owed direction; deadlock FALSE. **BLUE:** phase-driven stream-json wrapper, fired-record canary, causal provenance markers, invariant 7 — but shipped without its CHANGELOG/### BLUE blocks (process defect, repaired round 3 with catch-up entries).
- **Round 3 — RED FAIL: 17 gaps, mass 55.0.** Invariant 7 closes the round-2 watch; the residue shape is happy-path partiality — repairs holding for scheduled/completed runs but not the default mode (R3-2 medium-high), abort paths, or sibling surfaces. **LEAD:** all-successor docket carried with direction (the pass functioned as direction-setting; noted as protocol friction); deadlock FALSE. **BLUE:** all 17 addressed; rung 0 routed through the same wrapper via `--manual`; invariant 8 added; two blue leaf verifications of record (the `git log --output` gadget reproduced; the permissions doc re-fetched with a BROADER carve-out set than red's gap summary).
- **Round 4 — RED FAIL: 16 gaps, mass 46.0.** Full report re-read whole by all six lenses; zero external drift. Round-3 repairs closed the named path but not its sibling (`git format-patch -o` escapes, R4-2 medium-high; the payload-carrying `/self-improve` command, R4-1). **LEAD:** carried with the structural-close direction. **BLUE:** bare `Bash` deny at the tool boundary; git hook read-ALLOWLIST; `/self-improve` thin trampoline; R4-9 root invariant (no status timer-free); R4-11 probe attribution re-scoped honest.
- **Round 5 — RED FAIL: 10 gaps, mass 30.0, max severity medium — still converging.** Zero refuted external leaves; ten living-source claim sets zero-drift; the attached-form gadget reproduced at lens, merge, AND judge seats. Thread: round 4's structural closes were derived over the top-level session, not the surfaces where nightly work executes (seat population R5-1; corroboration pool R5-3). **LEAD:** all 10 carried with owed direction (risk_accepted unavailable across the board — every fix is a composing decision/predicate/clause); cross-cutting direction: state the invariant once over the TRUE surface; **deadlock FALSE — the debate continues into round 6.** **BLUE:** all 10 addressed — seat horn picked with the inheritance-independent hook close (R5-1), positive corroboration pool promoting OQ24 to built (R5-3), the cheap fan absorbed; claim_count 155 → 161.
- **Termination.** The 5-round safety ceiling fired before a round-6 red pass could audit the round-5 revision. No red PASS was ever issued; deadlock was never TRUE. Assembly stamps **UNVERIFIED** with the dispositions recorded in the Verdict-detail section above.

## Open questions carried past this run

Blue's final envelope open_questions, verbatim:

1. "OQ23(e) (new, R5-1): do debate.js-spawned Workflow seat agents inherit the headless --settings layer-1 profile? Build-time probe; the seat-surface Bash close does not depend on it (hook Bash-channel deny is the inheritance-independent layer), but the answer decides whether layer 1 ALSO binds seats or the hook is the sole seat-surface layer."
2. "OQ18(c)(3) (new, R5-5): the hook read-allowlist's handling of ATTACHED short-form flags (git format-patch -o<path>) must be leaf-tested, not assumed — the belt is declared known-incomplete on this class and the hook is the close of record."
3. "OQ1/OQ2: the bare --plugin-dir headless recipe and headless-specific hook-fence liveness remain the load-bearing Phase-4 acceptance probes (unchanged)."
4. "OQ22 (mid-drive --json-schema under stream-json) and OQ19/OQ11 (rung-2 Desktop per-task permission/--settings passthrough) remain build-time."
5. "Standing cross-corpus friction: the pin's plans/claude-port-plan.md path does not exist at 7bc501e (PINNED.md asserts a nonexistent path); port-plan quotes verified snapshot-grade against the working tree, not pin-grade."

The full 20+ item open-question ledger (OQ1–OQ24 with per-question status) lives in §8 of the embedded blue report. Candidates feed back to /research or /self-improve.

## Footnotes

The consolidated footnote namespace is embedded IN FULL above, in the blue report's closing section ("Footnotes (merged namespace; lane provenance noted per label)") — semantic word-based labels with per-lane provenance, source URLs, and access dates (2026-07-17 throughout). It is not duplicated here; this section records the assembly-level source-volatility notes so drift stays catchable:

- **[^MissedRun] re-pointed (round 5):** the round-1-era learn.microsoft.com Task Scheduler settings URL now 404s and the quoted MMC checkbox string has no stable doc leaf; semantics re-verified at the `TaskSettings.StartWhenAvailable` API property page, where the footnote now points.
- **#68375 (MCP headless):** OPEN, content zero-drift at round 5, but now carries a GitHub `stale` label — bot auto-close is a live drift risk; re-check at build.
- **#837 close-reason caveat:** GitHub issue close REASONS are not rendered in WebFetch markdown conversion; the COMPLETED-vs-NOT_PLANNED verification leans on round-1 `gh` CLI leaves (gh was classifier-blocked from lens seats rounds 2–5).
- **Port-plan citations:** snapshot-grade at working tree `6df52af`, NOT pin-grade (R1-7, risk_accepted run-scope — see Verdict detail).
- **Living sources last verified:** Claude Code docs (permissions, hooks, headless, scheduled-tasks, routines, web-sandbox, GHA-schedule), the pricing grid, and the GitHub issue set (#22055/#6631/#25621/#76239/#68375/#837/#14246) all re-fetched live 2026-07-17 (round 5), zero refuted leaves; the leaf-verification record per pair is `red/citation-ledger.md`.
