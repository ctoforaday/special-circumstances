# round 3 — lens 1 (leaf-node citation verification), instance 1 of 4

**Slice:** preamble, §0, §1 (even 4-way division of preamble+§0–§8: [pre/§0/§1] [§2/§3]
[§4/§5] [§6/§7/§8]) + the footnote definitions my sections reference: AlertFatigue,
Dependabot, DependabotFatigue, Goodhart, SelfCorrect, Reflexion, Voyager, DGM, DGMSakana,
SICA, STOP, FrictionRun3, FrictionRun4, Backlog, IdeasCorpus, EffReport, EfficiencyPlan,
ResearchCommand, HeadlessProbe, PortPlan, RedPatterns.

**Method:** full report re-read in 3 consecutive Read windows (1641 lines); citation-ledger
consulted first; round-2-changed text in my slice (§0 invariant 7 + R2-19 enumeration +
tree; §1.4 R2-11/R2-18; §1.5 R2-5/R2-6) audited fresh at the leaf; round-1/2 HIGH
verifications carried per the ledger rules (≤2 rounds elapsed, sections unchanged, immutable
pins). Ledger appended (round-3 lens-1 block).

## Verifications (corroboration HIGH unless noted)

1. **Preamble/§0 "invariant 7 ... per the lead's direction"** — debate.md round-2 ### LEAD,
   line 344: "derive the four fixes from it, and add it to §0's invariants, rather than
   patching gate-by-gate a second time." Corroborated. HIGH.
2. **§0 R2-19 enumeration totality** — printed tree recounted: 8 entries = 3 code artifacts
   + 2 command prompts + scheduling doc + skill file + manifest. Total. HIGH.
3. **§1.5 R2-5 stamp-at-creation feasibility** — "the bounded-pass sub-run dir, which the
   wrapper itself now creates when it runs setup-research-run.mjs wrapper-side": leaf-read
   the script (actual path
   `plugins/frank-exchange-of-views/skills/research-protocol/scripts/setup-research-run.mjs`;
   `git diff 7bc501e --` empty at that path). It takes `<runDir>` as argv and `mkdirSync`s
   the full skeleton — the wrapper choosing the runDir and stamping `inputs/.sleeper-origin`
   immediately after is buildable exactly as claimed. HIGH.
4. **§1.4 R2-18 arithmetic** — 30 × $0.10–0.50 = $3–15/mo (recomputed, correct); $50 cap
   headroom ≥3× holds at the band top; $2–5 × 30 = $60–150/mo (correct). Probe P2 $0.058
   consistent with §3.1 and [^HeadlessProbe]; the P2 figure itself stays MEDIUM
   (ephemeral-instrument, disposition-of-record: re-run at build — r1 ruling; no new
   attempt owed, the lane transcript was never committed so no extraction tool applies).
5. **§1.4 R2-11 graduation-queued** — internally consistent across §1.4/§2.3/§6 row 3. HIGH
   (internal-design claim; no external leaf).
6. **Carried HIGH without re-fetch** (per ledger rules): SelfCorrect, Reflexion, Voyager,
   DGM+DGMSakana, SICA, STOP (the "~page-of-code" paraphrase stays MEDIUM color per r2),
   Dependabot, DependabotFatigue, Goodhart (qualitative-only, blue-labeled), FrictionRun3,
   FrictionRun4, Backlog, IdeasCorpus, EffReport, EfficiencyPlan, ResearchCommand,
   RedPatterns, PortPlan (snapshot-grade; the pin-absent defect is standing and
   lead-docketed; AgentOrange working tree still 6df52af-clean this session).

## Findings

### L1-F1 — R2-5's run-window fallback is void for exactly the runs where it is needed (dead runs)

- **Location:** §1.5 (Tagging is by CAUSAL PROVENANCE) — "harvest ... additionally treats
  any run dir whose creation timestamp falls inside a wrapper-logged sleeper run window
  (ledger timestamps) as sleeper-origin even if markerless."
- **Evidence (internal cross-read, leaf-anchored):** the ledger record is written ONLY at
  step 7 (§2.2 step 7: "append cost JSON + stub path to the sleeper ledger"), and the
  design's own resume detection DEPENDS on its absence for incomplete runs (§2.2 step 0,
  R2-12: "started-but-unrecorded → resume"). A run that aborts and exhausts k=3 resumes
  (→ DEAD, §3.4) never reaches step 7: no ledger record, therefore NO logged run window —
  and the belt sweep (step-7 porcelain diff) never runs either. Both backstops die with the
  run. The wrapper-created dirs are still marker-stamped at creation, so the residual
  escape is a MODEL-authored run dir (Edit(<REPO>/research/**) is allowed) created
  in-session during a run that later dies — it enters next morning's harvest markerless,
  outside the origin cap, able to supply the "non-sleeper corroboration" R2-5 exists to
  deny. Dead runs are the design's own expected class (it builds DEAD/HALT machinery for
  them), and a dying run is precisely when anomalous session-authored artifacts are most
  likely.
- **Grade:** likelihood Low (needs model-authored dir + run death in the same night; no
  malice required — an injected or confused session suffices) × impact Medium (re-opens the
  R2-5 corroboration-supply hole for the abort path) × complexity-to-mitigate Low.
- **Required fix (cheap, two options, either suffices):** (a) log the run-window START
  timestamp in the operator-owned dir at step 0 (the start snapshot is already written
  there — give the fallback that timestamp pair, end = DEAD-mark time or next-fire time);
  and/or (b) run the origin-marker porcelain sweep at DEAD-marking time, not only step 7.
- **Pattern:** sibling-repair composition (R2-5's fallback composed against R2-12's
  "recorded = step-7 only" definition) / fail-open on the exceptional path.

### L1-F2 — §1.3 telemetry row is stale against the shipped reality this run itself demonstrates

- **Location:** §1.3 input-inventory table, board-telemetry.jsonl row — "(shipping in FEOV
  0.6.0 per the ratified efficiency plan)."
- **Evidence:** this run's OWN `trajectories/board-telemetry.jsonl` exists (checked by ls,
  2026-07-17) — the run executes under FEOV 0.7.0 and the telemetry line has shipped. The
  plan quote is faithful (plans/efficiency-phase.md does say 0.6.0), but the future tense
  tells a Phase-4 builder a harvest input is pending when it is already present in every
  current run dir. The fragment was graded MEDIUM by lens 1 round 1 and never repaired.
- **Grade:** likelihood n/a (text defect, confirmed) × impact Low (harvest.mjs design
  already lists the file as an input; only the availability framing is wrong) × complexity
  Low.
- **Required fix:** one clause — "(shipped as of FEOV 0.7.0 — present in this run's own
  trajectories/; specified by the ratified plan's PR-A.1)."
- **Pattern:** self-referential repo drift / stale-baseline.

### L1-F3 — R2-6's "every infrastructure class ALSO surfaces independently" is overbroad for intermittent failures

- **Location:** §1.5 (corroboration requirement) — "every infrastructure class in the
  bypass list ALSO surfaces independently on the doctor/dead-man line (§3.4), so the flag
  is not the sole channel."
- **Evidence (internal cross-read):** the §3.4 dead-man surface carries `last-successful-run`
  plus the LAST skip/abort reason — a single slot. An intermittent infra event (canary
  abort Monday, clean run Tuesday) is overwritten by the next success; by the time the
  human looks, the doctor line shows a healthy loop, and the `sleeper-only` docket flag IS
  the sole standing channel for that class. Intermittent flake is the modal infra failure
  on the design's own evidence (#76239/#68375 are both intermittency bugs). DEAD/HALT
  markers persist until cleared and are fine; the overbreadth is for the transient members
  of the bypass list (aborted-run, ledger-unparse skip, canary abort, hook-crash).
- **Grade:** likelihood Medium (intermittency is the common case) × impact Low (the flag +
  1-stub cap + provenance contract still bound the harm; only the "not the sole channel"
  argument loses a leg) × complexity Low.
- **Required fix:** either scope the sentence honestly ("surfaces on the dead-man line at
  least until the next successful run; persistent failures surface durably") or make the
  cheap mechanism true (rolling abort-reason history file beside the ledger; doctor prints
  count-in-last-30-days, not only last).
- **Pattern:** false-equivalence disjuncts (a persistent-marker channel and a
  last-value-overwritten channel argued as equivalent witnesses).

### L1-F4 — NOTE (no defect): [^AlertFatigue] upgrade is one citation swap away

- **Location:** §1.1 — "vendor analyses put the acted-upon fraction under one in five
  (figure seen at search-digest level only — NOT leaf-verified...)" + [^AlertFatigue].
- **MUST-TRY line:** fresh WebSearch this round (2026-07-17; r1's attempt found no
  primary). Found: (a) ACM Computing Surveys, "Alert Fatigue in Security Operations
  Centres: Research Challenges and Opportunities," doi 10.1145/3723158 — peer-reviewed
  survey of the phenomenon; (b) 2026 State of Production Reliability report (survey
  n=1,039 SRE/DevOps professionals, Feb 2026): "57% report that fewer than 30% of those
  alerts are actionable," "83% ignoring or dismissing alerts at least occasionally."
- **Disposition:** blue's self-grade (LOW number / MEDIUM phenomenon) is honest — no gap.
  But the unpinnable "under 1 in 5" can be REPLACED by the pinnable 2026 survey figure and
  the phenomenon upgraded toward HIGH by citing the CSUR paper. Offered as an optional
  round-3 upgrade, priced trivial.

## MUST-TRY / impossibility lines for every below-HIGH grade in slice

- [^AlertFatigue] number: WebSearch attempted this round (results above) — report's
  specific figure still unpinned; replacement figure found.
- [^HeadlessProbe] P2 figures MEDIUM: impossibility — ephemeral instrument (lane transcript
  never committed); no extraction path exists; disposition-of-record stands (re-run at
  build).
- [^PortPlan] MEDIUM/snapshot-grade: impossibility — path absent at pin 7bc501e (git-show
  verified r1); working-tree corroboration re-confirmed r2 and tree still clean at 6df52af;
  defect standing on the lead's docket.
- [^STOP] "~page-of-code" MEDIUM (color): attempted r2 at ar5iv — literal absent,
  paraphrase corroborated; standing.
- [^Goodhart] qualitative: no attempt owed — blue carries no figures and labels
  search-digest level; qualitative use verified r1 via Wikipedia leaf.

## Friction

- **Glob/Grep refuse the run directory** ("Path does not exist") because it sits outside
  the registered working directories, while Read and Bash reach it fine — every search in
  the audit surface forced a Bash grep/find spawn (10–100x cost). The lens seat's audit
  surface should be a registered working directory, or Glob/Grep should honor the same
  access set as Read.
- Minor: ls-with-trailing-slash on the run dir intermittently returned ENOENT under Git
  Bash while the same path without the slash listed fine — cost two diagnostic
  round-trips; no durable blocker.
