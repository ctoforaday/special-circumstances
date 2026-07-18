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

---

## 0. Design summary (the implementable shape)

```
plugins/sleeper-service/
├── .claude-plugin/plugin.json            # requires frank-exchange-of-views
├── skills/continuous-learning/SKILL.md   # promotion ladder: insight → MEMORY → idea stub →
│                                         #   graduated research → rules/skills (HUMAN-GATED);
│                                         #   expand-existing-before-append discipline
├── commands/
│   ├── self-improve.md                   # daily loop: preflight → harvest → pick ONE →
│   │                                     #   bounded research → idea stub. Writes ONLY
│   │                                     #   research/ + ideas/.
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
│   │                                     #   snapshot + step-7 auto-compare (R1-26/R2-7),
│   │                                     #   resume cap k=3 + per-cause HALT (R1-29/R2-10),
│   │                                     #   dead-man record (R1-29/R2-9)
│   └── sleeper-guard (PreToolUse hook)   # write-fence: deny writes outside research/ +
│                                         #   ideas/ for sleeper runs; fence LIVENESS proven
│                                         #   per run by the step-0 canary's hook-authored
│                                         #   fired-record (R2-2) — marker loss or hook
│                                         #   silence aborts the run instead of failing open
│                                         #   (R1-28); deny-set baked into the binary, not
│                                         #   read from editable config
└── docs/scheduling.md                    # the ladder: manual → OS scheduler + claude -p →
                                          #   Desktop scheduled task → cloud Routine → GitHub
                                          #   Actions; preflight + ceilings for every
                                          #   unattended rung
```

New code surface is deliberately small — exactly THREE new code artifacts (harvest.mjs,
the sleeper PreToolUse guard, and the scheduler wrapper; round 0 said TWO — the wrapper
was already implicit in §3.4's rung-1 row, and round 1 promotes it to a named artifact
because five gate-side controls now live in it: preflight, ledger, canary, snapshot,
resume cap) plus two command prompts, a scheduling doc, **the continuous-learning skill
file, and the plugin manifest** (round-2 totality fix, R2-19: the two latter entries are
printed in the tree above but were absent from round 1's enumeration, which also made
"everything else reuses shipped machinery" literally false for them — they are new PROSE
artifacts, not code, and are now enumerated); everything else
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
   aborts on absence (R2-1/R2-2); the step-7 snapshot compare raises a flag and fails the
   next preflight closed on delta (R2-7); the preflight verifies the executing plugin
   copy's hash against the operator-approved value and aborts on mismatch (R2-8); the
   dead-man/HALT records carry reasons a doctor line and the session-start warning print
   (R2-9/R2-10/R2-18). The checked-record discipline is the design's answer to
   round 2's board, stated once so future gates inherit it instead of re-learning it
   gate-by-gate.

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
| `research/*/trajectories/board-telemetry.jsonl` | round-by-round board profile, mass trend, new-mint profile (shipping in FEOV 0.6.0 per the ratified efficiency plan)[^EfficiencyPlan] | debate-engine tuning; termination evidence |
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
lane-1/adversarial — the tie-break]. Tunable constants live in the script. The formula is
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
`/self-improve`, same code path, zero standing cost — is the DEFAULT and may be
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
the run window and lacks a marker; harvest.mjs reads ONLY the marker, never the dir name,
and additionally treats any run dir whose creation timestamp falls inside a wrapper-logged
sleeper run window (ledger timestamps) as sleeper-origin even if markerless. Harvest tags
every row from a marked dir `origin: sleeper` and **caps sleeper-origin recurrence
contribution at 1 per class** regardless of how many nights repeat it.

**The corroboration requirement is DECIDED, and each horn's cost owned (round-2 repair,
R2-6 — the round-1 "requires non-sleeper occurrence — or, failing that, flags
`sleeper-only`" was two materially different mechanisms wearing one sentence: a hard bar
that structurally blinds the loop to failure classes observable only in its own
headless/--bare/MCP path, or a label whose containment reduces to the human noticing a
flag at the design's own Dependabot base rate).** The decision is the severity-gated
bypass (red's proposed shape, lead-endorsed, adopted): **infrastructure-failure classes**
— aborted-run, fail-closed events (ledger-unparse skip, canary abort, plugin-hash
mismatch), hook-crash, DEAD/HALT markers — enter the docket flagged `sleeper-only`
WITHOUT corroboration, because these classes are by construction observable only in
sleeper runs, and a hard bar would blind the loop to its most-relevant failures precisely
as the corpus becomes sleeper-dominated; **ordinary capability/protocol classes** require
at least one non-sleeper occurrence before entering the top-3 comparison. Costs owned in
text: for ordinary classes, a sleeper-only phenomenon waits for human-present
corroboration (accepted — those venues overlap human-present runs, and the docket still
lists the class below the fold, visibly); for infrastructure classes, containment does
lean on the human seeing the `sleeper-only` provenance flag — bounded by the 1-stub/run
cap, the mandatory provenance contract, and the fact that every infrastructure class in
the bypass list ALSO surfaces independently on the doctor/dead-man line (§3.4), so the
flag is not the sole channel. §6 row 10's grade tracks this decision.

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
   staging (docket + red-memory mirror); start snapshot (git porcelain + guardrail
   hashes → operator-owned log; compared at step 7 — R2-7).
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
   paragraph. The pick returns to the wrapper as structured output (the phase drive's
   `--json-schema` leg), because step 4's staging needs it wrapper-side.
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
   (re-scoped round 2). Verdict will honestly be UNVERIFIED — correct for a stub.
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
   end snapshot AND COMPARE (R2-7): end-vs-start and start-vs-prior-run's-end guardrail
   hashes — on any mismatch, write a dated TAMPER flag beside the ledger, surface it on
   the doctor line, and fail the NEXT preflight closed until the human clears it (the
   window now has a watchman: the wrapper is the comparer, the doctor line + blocked
   preflight are the human surface — invariant 7); origin-marker sweep of the porcelain
   delta (R2-5); update last-successful-run + skip/abort reason for the dead-man surface
   (R1-29/R2-18); exit with the JSON envelope. [minority: lane-3/local-probe — the
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
                    where the full MCP surface is present>
est-cost-to-graduate: <keeper-run + build estimate> [minority: lane-3/local-probe]
graduation_trigger: <recurrence/severity condition under which /graduate is worth a human's
                    time> [minority: lane-1/adversarial]
status:            <open | stale | graduated | rejected — "open" at mint; harvest
                    auto-stales at 30 days computed from the dated filename (R1-22);
                    humans set the terminal states>
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
  `--permission-mode` (incl. `dontAsk` — auto-denies anything not allow-listed; documented
  for "locked-down CI runs").[^CliReference][^HeadlessDocs]
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
   #837).[^SlashHeadlessIssues] The current doc's statement supersedes them, but the port
   plan's Phase-4 verify step ("Headless `claude -p \"/self-improve\"` produces a run dir +
   idea stub; touches only research/+ideas/")[^PortPlan] must remain the acceptance test —
   doc-says is not run-verified (probe P2 covers the non-bare case on 2.1.212; the
   bare+--plugin-dir case is still open). **scheduling.md ships a preflight** (trivial
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
| 0 | Manual `/self-improve` in a session | always available; the default until the human opts into scheduling (port plan resolved decision 6: daily cadence, scheduling always human-opt-in)[^PortPlan] |
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
the k=3 cap bounds per-DIR, not per-CAUSE — a deterministic root cause that survives a
fresh dir would otherwise burn up to k×$5/night until the monthly cap trips (~3 nights
at the ceiling) and again after the cap resets. **Per-cause HALT (R2-10):** the wrapper
normalizes each death's abort reason (exit class + first abort-record line) into a dead
SIGNATURE recorded beside the DEAD marker; after M=3 consecutive fresh-dir deaths with
the same signature the wrapper writes a `HALT` marker — preflight refuses to launch
until the human clears it (§2.2 step 0) — and raises the dead-man flag with the
signature as the printed reason. **Dead-man surface (R1-29; re-argued round 2, R2-9):**
"loudly" gets a defined reader — the wrapper maintains `last-successful-run` plus the
last skip/abort REASON (R2-18) in the
operator-owned dir, and `/prosthetic-conscience:doctor` reports "last successful sleeper
run: N days ago (last skip reason: <reason>)". Round-2 honesty (R2-9): doctor is
PULL-only, and the loop's adoption motive removes exactly the babysitting that made "the
surface the human already looks at" an assertion rather than evidence. Decision, both
horns owned: (a) a cheap push-adjacent surface is added — the same staleness check runs
as a SessionStart-hook warning line in every INTERACTIVE session on this box once
`last-successful-run` exceeds N=7 days (one file-stat; a channel the operator passively
receives whenever they use Claude Code at all — invariant 7's human surface — and one
that is independent of the sleeper scheduler, so scheduler death does not silence it);
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
machinery bind rung 1 fully and rung 2 at most partially; at rungs 3–4 most layers do not
exist unless rebuilt for that environment. **Climbing past rung 2 is itself a
graduation-grade decision** (its own idea stub, its own human gate) — never a config
toggle: as documented round 0, an operator climbing the ladder would reach the
maximum-blast-radius configuration (fully autonomous, connectors default-included, none of
§4.2) still believing the gates were on.

| Gate / control | R0 manual | R1 OS + `-p` | R2 Desktop task | R3 Cloud Routine | R4 GitHub Actions |
|---|---|---|---|---|---|
| L1 permission profile (operator-owned `--settings`) | session rules | **YES** | PARTIAL (per-task permission config; `--settings` passthrough unverified — OQ11/OQ19) | **NO** — no permission prompts; the profile file is absent from the fresh clone, and committing it moves it inside the tree the loop's PRs touch | PARTIAL — a workflow-committed settings file sits inside the PR-touchable tree (rung-4 inversion, R1-27) |
| L2 hook fence + step-0 denial canary | SPLIT (R2-16): fence YES (cache copy); canary n/a — no wrapper at rung 0, so fence liveness is UNVERIFIED per-run when manual | **YES** (wrapper canary + fired-record — R2-2) | PARTIAL (canary needs the wrapper) | only if hooks load from committed plugin config; canary absent | same as R3 |
| Wrapper controls (preflight, ledger, idempotence, resume cap, snapshots) | n/a (human present) | **YES** | NO (Desktop owns scheduling; wrapper absent unless the task invokes it) | **NO** | PARTIAL (rebuildable as workflow steps) |
| L5 git/PR gate (server-side, branch protection + CODEOWNERS) | YES | YES | YES | YES — plus native `claude/`-branch push restriction | YES |
| L6 `disable-model-invocation` on /graduate | YES | YES | YES | YES (doc-verified for scheduled fires) | YES |
| Per-run budget (`--max-budget-usd`) | YES | **YES** | UNKNOWN (per-task config — OQ19) | platform caps only (subscription draw-down + daily routine cap) | YES (flag in workflow) |

Reading: rung 1 is the only rung where every layer is live as designed — which is *why* it
is the recommended default among scheduled rungs, not merely a convenience preference.
Rungs 3–4 keep the
load-bearing layer (5) and shed most of the rest; adopting them means re-deriving the
stack for that environment, documented in the adoption stub. **Rung-0 ledger scope,
declared (R2-16b):** the monthly ledger is wrapper-written, and rung-0 manual runs do not
pass through the wrapper — manual-session spend is OUT-OF-LEDGER BY DESIGN. Consequences
owned: an operator who later opts into scheduling starts the ledger blind to prior manual
spend, and a mixed manual+scheduled month undercounts at the ledger — accepted because
rung-0 spend is human-present (visible at invocation, billed to an attended session) and
the cap's job is bounding UNATTENDED spend; if mixed usage becomes routine, the cheap
upgrade is invoking manual runs through the same wrapper (the code path already exists).

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
      "Bash(git status)", "Bash(git diff)", "Bash(git log *)",  // pinned argv; no node, no model-chosen write targets (R1-16)
      "Workflow",                                               // debate.js via the harness Workflow tool (R2-3); scriptPath
                                                                //   pinned to the read-only plugin copy by the sleeper-guard
                                                                //   hook (native rule-level scoping is OQ20)
      "mcp__qmd__*", "WebFetch(domain:arxiv.org)"               // extend deliberately, per need; NO WebSearch nightly (R1-18)
    ],
    "deny": [
      "Edit(<REPO>/.claude/**)", "Edit(<REPO>/plugins/**)", "Edit(<REPO>/CLAUDE.md)",
      "Edit(<REPO>/MEMORY.md)", "Edit(<REPO>/.mcp.json)", "Edit(<REPO>/.github/**)",
      "Edit(<REPO>/plans/**)", "Edit(<REPO>/projects/**)",
      "Edit(**/hooks.json)", "Edit(**/settings*.json)",
      "Read(//c/Users/gbloc/.claude/settings*.json)",   // belt: Read deny also blocks Edit (v2.1.208) — R1-17
      "Read(//c/Users/gbloc/.claude/.credentials*)",
      "Read(//c/Users/gbloc/.claude/projects/**)",      // session transcripts: the high-value exfil target
      "Edit(//c/Users/gbloc/.claude/**)",
      "Bash(git push *)", "Bash(git commit *)"
    ],
    "disableBypassPermissionsMode": "disable"           // INSIDE permissions — R1-8
  }
}
```

(Primary read protection is the allow scoping itself — under `dontAsk`, Read outside the
allow set is auto-denied; the `Read(...)` deny lines are belt on the named high-value
targets. `disableAutoMode`, the same escape-hatch class as
`disableBypassPermissionsMode`, is a candidate addition — its key name/scope must be
leaf-verified against the current permissions doc at build before this report asserts it:
open question 17.)

Why this shape is sound rather than merely asserted:

- `dontAsk` **auto-denies anything not in the allow rules** — the default is closed, so a
  channel we forgot is denied. This is exactly the allowlist inversion red's
  invariant-soundness pattern demands (a denylist alone leaves new tools
  default-open).[^PermissionsDoc][^RedPatterns] The deny list is belt-and-suspenders on top.
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
| 2. Path-fence hook | `sleeper-guard` PreToolUse (Go, sc-quality-gate class — hooks verifiably fire on subagent writes[^IdeasCorpus]) blocks Edit/Write/NotebookEdit/Bash-write targets outside research/+ideas/ when the sleeper marker identifies the run (env var set by the wrapper + `agent_id`/`agent_type` fields[^HooksDoc]); fence LIVENESS is proven per run by the wrapper's step-0 canary drive, round-2 mechanism (R2-1/R2-2 — the round-1 deny-outcome check could not isolate this layer even in principle: layers 1 and 2 fence the IDENTICAL boundary, so every canary write is denied by layer 1 whenever it is live, and on marker loss — the modal drift case — the fence is dormant, the canary would PASS, and the run would proceed under exactly H4's refuted permissions-only configuration while stamping "enforcement proven"). The repaired mechanism, per invariant 7: the guard, on every sleeper-fence DENY it issues, appends a fired-record line (tool, target path, nonce echo, decision) to the operator-owned dir — hooks are harness-side processes, not model tool calls, so the record is reachable by the guard and unforgeable by the session (no Bash beyond pinned git argv, Edit denied, Workflow scriptPath pinned). The wrapper's phase-0 canary (nonce-named target — §2.2 step 0) then requires BOTH (a) the write was denied in the event stream AND (b) the fired-record contains the nonce with decision=deny — proving THIS layer fired, not merely layer 1. Marker loss ⇒ the guard never engages ⇒ no fired-record ⇒ ABORT, even though layer 1 still denies the write: fail-closed restored, and OQ2's fence-dormant acceptance test now passes as designed. (Belt, verify-at-build: if stream-json deny events carry distinguishable PROVENANCE — hook-block message vs permission-rule denial — the wrapper parses it as a second witness; OQ21.) Marker-loss base rate (scheduler misconfig, shell nesting, rung migration) still exceeds any adversarial term in §6 (R1-28). Deny-set **compiled into the binary**, not read from editable config [minority: lane-1/adversarial — compiled-in deny-set]. Hook semantics favorable: "A blocking hook also takes precedence over allow rules," "Hook decisions don't bypass permission rules" — layers 1+2 can only tighten each other, never loosen[^HooksDoc][^PermissionsDoc] | Loop modifies hook source in `plugins/` in-repo → the repo copy is not the executing copy — now PINNED, not assumed (R1-15): non-bare runs execute the plugin cache under `~/.claude/plugins/`; the `--bare` recipe's `--plugin-dir` points at the operator-owned read-only copy (`~/.claude/sleeper/plugins/`), NEVER the working tree (round 0 left this unreconciled — under the preferred recipe as written, the repo copy could have BEEN the executing copy, voiding this layer's independence leg). `Edit(plugins/**)` is denied anyway, and a modified repo copy reaches either executing copy only through the human-gated plugin-update path |
| 3. Self-protection of the guardrail definitions | Every file that *defines* layers 1–5 (permission profile, plugin hooks, scheduled-task definition, the command markdown of /self-improve and /graduate, the guard binary) is outside the write boundary; the hook additionally denies writes to the plugin cache, settings, hooks.json, and its own binary path. The empty-bin bootstrap guard keeps hook failure loud[^HooksJson][^Backlog], and the wrapper aborts if the guard binary is missing (fail-closed, not no-op) [minority: lane-1/adversarial — the fail-closed abort]. The loop can propose changes to its own guardrails only as **idea stubs** walking the same human-gated graduation path — closing the DGM failure mode by construction rather than instruction[^DGMSakana] | Loop disables hooks by editing settings → settings paths are inside the deny-set; residually, chmod-readonly on guard files as defense-in-depth — a DESIGN-PROPOSED measure of this report's own, not community-sourced (corrected round 1, R1-13: the #22055 thread documents the PreToolUse protected-files workaround verbatim; chmod-444 is not in the thread) [minority: lane-1/adversarial][^PermAskBypass] |
| 4. Subprocess residual (the honest gap) | File rules "don't apply to arbitrary subprocesses that read or write files indirectly, like a Python or Node script that opens files itself"; OS-level enforcement requires the sandbox.[^PermissionsDoc] Under `dontAsk` the channel is nearly closed (Bash deny-by-default except the enumerated allowlist), but a named script is itself a subprocess | In order of preference: (a) enable Claude Code **sandboxing** for the scheduled profile so filesystem writes outside the run tree are OS-blocked for Bash and children (permissions + sandbox are the documented defense-in-depth pair); (b) where the sandbox is unavailable (current Windows box), the round-1 shape (R1-16) removes all MODEL-AUTHORED script execution from the session — re-scoped round 2 (R2-3: the round-1 "removes ALL script execution" was falsified by the design's own step 4, which invokes FEOV nightly, and the shipped command's execution locus is MIXED per probe[^ResearchCommand]: setup/capture are session-Bash node calls, the debate engine is a **Workflow tool** invocation). The true premise, stated: (i) the two Bash node scripts are re-hosted WRAPPER-SIDE (§2.2 steps 0/4/7), so the in-session Bash allowlist stays pinned-argv read-only git commands only (no rule grants the model argv that chooses a subprocess write target; the round-0 `Bash(node scripts/*)` grant contradicted "read-only analysis tools" in-document AND mismatched the printed layout); (ii) the ONE remaining in-session script channel is the Workflow tool running debate.js — PINNED CODE from the read-only operator-owned plugin copy, model-influenced only through schema-bound args (topic/runDir/lanes/models), with the `Workflow` tool added to the §4.2 allow list and the sleeper-guard hook additionally matching Workflow calls and denying any scriptPath outside the plugin copy (PreToolUse fires on every tool; whether permission RULES can scope Workflow's scriptPath natively is OQ20); (iii) the workflow's SEAT AGENTS are full permission-engine + hook subjects — leaf evidence, not assumption: sc-quality-gate verifiably fired on workflow-agent writes and red-auditor wrote its memory under hook coverage[^IdeasCorpus]; (iv) the workflow SCRIPT's own Node fs writes are a layer-4-class subprocess residual, bounded by fixed pinned code + the wrapper's step-7 snapshot compare (R2-7). Compound-command/redirection/traversal matching for the surviving git rules is a named build-PR test (OQ18); plus a hook clause blocking Bash whose command text targets the guardrail set ("a blocking hook … stops the tool call before permission rules are evaluated"[^PermissionsDoc]); and layer 5's tamper evidence is wrapper-snapshot-backed (R1-26), no longer working-tree-only. Residual RISK-ACCEPTED (§6 row 4, re-argued round 1 WITHOUT the benign-actor premise — R1-21) and stated honestly in docs/scheduling.md rather than papered over |
| 5. Git as the promotion gate (the load-bearing layer) | The loop never commits and never pushes (`git commit`/`git push` denied — lane-2 denies both; lane-3 denies push and notes sc-push-freeze-guard warns-only by design, so the loop profile upgrades it to a hard deny for loop sessions[^PushGuard]). The loop's entire output is an uncommitted working-tree delta under research/+ideas/ that the operator reviews next session — `git status`/`git diff` **is** the review surface, and anything outside the boundary shows up as an anomaly [minority: lane-2/primary-literature — the uncommitted-delta review model]. Round-1 hardening (R1-26): an uncommitted delta is SELF-ERASING for exactly the residual it backstops — the layer-4 subprocess channel that can write out-of-boundary can also revert before review, and run N+1 fires before run N is reviewed; DGM's lineage was durable and externally inspectable, and an uncommitted working tree is its opposite. So the wrapper takes an out-of-session snapshot (`git status --porcelain` + SHA-256 hashes of the guardrail file set) at run START and END, appended to an operator-owned log outside the repo: a write-then-revert inside the run shows as a start/end delta the model cannot reach. **The delta now has a WATCHMAN (round 2, R2-7 — a window nobody reads is telemetry, not detection; round 1 reintroduced for tamper evidence the exact undefined-receiver defect R1-29 had fixed for loop death):** at step 7 the wrapper itself COMPARES end-vs-start and start-vs-prior-run's-end guardrail hashes; on any mismatch it writes a dated TAMPER flag beside the ledger, the doctor line prints it, and the NEXT preflight fails closed until the human clears the flag — so run N+1 cannot fire past unexamined tamper evidence, and detection latency is one run, not one human review cadence (invariant 7). This restores the durable tamper-evidence property the DGM analogy actually requires. (Heavier alternative if snapshot review proves insufficient: wrapper auto-commit of the run delta to a `sleeper/` branch — commit denial binds the model, not the wrapper.) Optionally the run operates on a `sleeper/`-prefixed branch or dedicated worktree. **Server-side out-of-band enforcement** [minority: lane-1/adversarial]: promotion into rules/skills happens ONLY via PR into main under branch protection + CODEOWNERS on `plugins/**` and `.claude/**` — controls no local write can alter. Cloud routines get this natively: pushes restricted to `claude/`-prefixed branches by default[^RoutinesDocs] | The loop cannot route around GitHub's server-side enforcement from a local session; the erosion path left is *social* (human rubber-stamping PRs), which is a review-culture problem no mechanism solves — named honestly in scheduling.md [minority: lane-1/adversarial]. A purely local repo downgrades to layers 1–4 + review discipline — accepted because the suite's operating model already assumes a GitHub-hosted repo with PR review (PR #1 flow) |
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
| bounded research pass (daily) | bulk-tier lanes; judgment inherit-session NOT granted headless — pin judgment to bulk in daily runs and accept the grade | smoke-scale (~50k tokens measured)[^Backlog]; the daily run produces *stubs*, not verdicts; its output is always re-judged at graduation |
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
input ~90%. The Fable/Mythos/Sonnet-5 tokenizer counts ~+30% more tokens than legacy
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
| 1 | Headless plugin-command expansion regresses (#837 class) or `--bare` becomes the `-p` default and the nightly run silently loses plugins | Low–Medium (doc explicit today; flip announced) | High (loop dead) | Low | Mitigate: step-0 loaded-plugins assertion aborts loudly; Phase-4 verify step is the acceptance test; version pin; idempotent scheduler makes failure loud-but-harmless |
| 2 | Daily stubs too shallow to graduate (H2 falsifier) | Medium | Medium | Low | Measurable via ledger stub→graduation survival rate; named revisit trigger: switch to weekly-full-strength |
| 3 | Proposal fatigue at the human gate | Medium (Dependabot base rate) [minority: lane-2/primary-literature] | Medium | Low | 1-stub/run cap + dedupe against open stubs younger than the staleness window; stubs auto-stale at 30 days via harvest's filename-date sweep and re-enter the docket flagged `stub-stale` (mechanism specified §1.4/§2.3 round 1 — R1-22; round 0's "age out visibly" was policy-without-mechanism); cadence revisit trigger in §1.4 (R1-14) |
| 4 | Subprocess write outside boundary (layer-4 residual; no full sandbox on Windows) | Low-Medium (re-graded round 2, R2-3: round 1's Low rested on "no script execution" — false as printed; the true surface is pinned git argv PLUS the Workflow channel running pinned debate.js) | High (boundary breach) | Medium | Layered (allowlist + hook + wrapper snapshots); residual RISK-ACCEPTED, re-argued round 2 on the TRUE premise (R2-3, superseding the R1-21 re-argument whose leg (a) the round-1 text falsified) and still WITHOUT actor benignity: (a) channel narrowness, restated — the subprocess surface is the pinned read-only git argv set plus the Workflow tool executing PINNED code from the read-only plugin copy under schema-bound args (hook-enforced scriptPath — §4.3 layer 4; seat agents inside the workflow are permission+hook subjects, leaf-verified[^IdeasCorpus]); a breach therefore needs a permission-engine bug AND a write gadget in either pinned-argv git or pinned debate.js code simultaneously; (b) detection outside the model's reach — wrapper step-7 snapshot COMPARE with TAMPER flag + fail-closed next preflight (R2-7: the delta now has a defined watchman, so this leg is backed, not presupposed); (c) L×I×Cx — full-sandbox complexity on Windows exceeds the bounded likelihood × impact today. No premise about actor intent is used. Revisit triggers: any hook-log blocked guardrail-targeting Bash or Workflow call; any TAMPER flag (now wrapper-raised and preflight-blocking, R2-7) [minority: lane-3/local-probe — the trigger] |
| 5 | No programmatic quota pre-check | Certain (no spend-limit API; rate-limit API unreachable at this auth tier — §5.1/R1-9, cell requalified round 2, R2-13) | Low (static ceilings suffice) | Low | RISK-ACCEPT with §5.1's layered static guards — a Console-scraping workaround would be complexity strictly worse than the risk |
| 6 | Scheduler misses days (sleep/laptop) | High | Trivial | Low | RISK-ACCEPT by design: missed-run-tolerant trigger + idempotence; a skipped day costs nothing |
| 7 | Loop self-modifies its guardrails | Low per-attempt, compounding | Critical | Medium | Layers 1–7 (§4.3); the STOP/DGM/AI-Scientist evidence says treat per-attempt probability as nonzero forever |
| 8 | MCP headless flake (open bugs #76239/#68375; #32191 CLOSED as duplicate, kept as phenomenon history — corrected R1-5) starves qmd recall | Medium | Low | Low | Daemon `/health` preflight + strict-mcp-config + Grep/Read degradation (proven workable) [minority: lane-3/local-probe] |
| 9 | Version-bump empty-bin window: nightly run during the update dance hits a hook crash-storm | Medium | Low | Low | Bootstrap guard already shipped in hooks.json[^HooksJson]; scheduler preflight can run doctor check; accept residual noise [minority: lane-3/local-probe] |
| 10 | Self-poisoning of the ranking input — edit AND authorship paths (re-scoped round 1, R1-25; mechanism rebuilt round 2, R2-5/R2-6) | Low-Medium (round 2: the round-1 glob-keyed tag was circumvented by the default nightly control flow — the FEOV sub-run dir escaped it; the causal-provenance marker closes that hole, so the grade holds on the new mechanism, not the broken one) | Medium | Low | Stub-files-only write pattern for the edit path; for the authorship path: wrapper-stamped origin markers on EVERY spawned run dir (sub-runs included) read by harvest — never dir names — with recurrence capped at 1 per class (§1.5, R2-5); corroboration is DECIDED severity-gated (R2-6): infrastructure-failure classes enter flagged `sleeper-only` without corroboration (also independently visible on the doctor/dead-man line), ordinary classes require one non-sleeper occurrence; the grade tracks that decision — residual = flag-reliance for infra classes + corroboration-wait for ordinary sleeper-only phenomena, both owned in §1.5 [minority: lane-3/local-probe] |
| 11 | Cloud Routine acts through the operator's identity with all connectors default-included | Medium | Medium | Low | If the cloud alternate is used: trim connectors, keep `claude/`-branch restriction ON; document in scheduling.md [minority: lane-3/local-probe] |
| 12 | Unattended overspend | Low | Medium | Low | Per-run `--max-budget-usd` + `--max-turns` + ledger preflight (operator-owned, fail-closed — R1-19); platform caps on cloud |
| 13 | Exfiltration via read+egress: an injected loop reads box-local secrets and encodes them into its egress (web queries, fetch URLs) — added round 1, R1-17 | Low-Medium | Medium-High | Low | Mitigate: Read/Grep/Glob allow-scoped to the repo (dontAsk auto-denies all other reads); belt Read-denies on `~/.claude` settings/credentials/session transcripts; WebSearch removed from the nightly profile. Residual egress = arXiv-domain WebFetch URLs + qmd queries (local) — RISK-ACCEPTED with rationale: single-operator box, repo largely public, the remaining channel is narrow, enumerable, and session-logged |
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
  CHANGELOG Round 2.

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
    is untested.
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
    `Bash(git ...)` allow rules (R1-16c) — named build-PR test.
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
  settings above command-line arguments.
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
  https://github.com/anthropics/claude-code/issues/68375; #32191 (`-p` with HTTP MCP server
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
  possible after a scheduled start is missed" (Microsoft Task Scheduler settings docs).
  Surveyed 2026-07-17.
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
  Fable/Mythos/Sonnet-5 tokenizer counts ~+30% more tokens than legacy counting. The ≤24h
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
  research/+ideas/"; resolved decision 6: daily cadence, scheduling always human-opt-in.
