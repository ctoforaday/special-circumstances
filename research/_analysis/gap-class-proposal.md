# Proposal: gap classes — track problems, not occurrences

Status: DRAFT (user-originated, 2026-07-17 during run 5). Queued for ideas/backlog.md
at capture; candidate for folding into the record-layer plans-PR (board.jsonl).
Debate-gated like the record layer — this changes red's bookkeeping semantics.

## The threat model this closes

The shell game: blue "agrees" to a fix, edits the anchored text, and the problem
reappears as one or more smaller or different findings under fresh ids. Per-instance
accounting scores this as convergence (1 closed, N minted, each graded smaller), so
reclassification dodges tracking. Run 5's integrity audit (2026-07-17) found no such
conduct — but established that detection relied on judge ancestor-reads and agent
honesty, not structure. Three structural facts make the dodge cheap today:

1. Mass is summed per instance; a lineage that splits can shrink the board total
   while multiplying loci (R1-25 → five descendants was the legitimate instance of
   the exact arithmetic a shell game would produce).
2. Supersedes chains join VERTICALLY (one lineage) but nothing joins HORIZONTALLY
   (same kind of defect, different chains). "Asserted composition presented as
   documented" occurred ≥2× in run 5 under unrelated ids (R3-1; the R2-3 caveat) —
   invisible as a recurrence.
3. Gaps anchor to text loci, so editing the locus is always a valid "fix" even when
   it relocates the problem. Point-anchored tracking cannot distinguish resolution
   from relocation. Within-run class tracking exists nowhere (gap-patterns memory is
   cross-run only).

## Mechanism

**Class field on every gap.** Red-merge assigns each minted gap a `class` slug from
an append-only in-run registry (`red/classes.md` or board.jsonl rows): slug, one-line
definition, first-instance id. The near-match rule extends: before minting, check
the class registry, not just open ids. Cross-run seed: gap-patterns memory slugs
become the initial vocabulary, closing the loop with red's existing memory.

**Class-level accounting in telemetry.** Per round, per class: instance count, open
mass, RECURRENCE count (instances minted after the class first had zero open
instances). Class severity = max(instance severities) with a floor that rises on
recurrence — many small instances cannot dilute a class below attention.

**Forced generator-fix escalation.** Nth recurrence of a class (N=2 proposed)
auto-dockets a CLASS review: judge directs blue to address the generator (a template,
a rule, a habit of the drafting process) and the class cannot close on instance fixes
alone — it closes when a named generator change ships + one full round passes with
zero new instances. This converts "fix the sentence" into "fix what keeps producing
such sentences."

**Anti-shell-game property.** Splitting or reclassifying an instance cannot escape:
successors inherit the ancestor's class by default (merge may re-class WITH a
recorded reason — same discipline as grade reconciliations), so a split INCREASES
the class's instance count and trips the recurrence escalator instead of resetting
attention. Class mass is reported beside board mass; a decaying board with a growing
class is now a visible contradiction.

## Gaming the classifier (and mitigations)

Classification is itself gameable in both directions: singleton-everything (every
gap its own class → no-op) and umbrella-everything (one mega-class → noise).
Mitigations: registry is append-only with definitions (creating a class is a written
act the judge can review); capture audit reports singleton-class rate and largest
class share as red flags; judge dockets include class assignment on contested gaps;
the terminal judge reviews the registry once at run end (cheap — it is a slug list).

## Fit with existing bones

- board.jsonl (record-layer proposal): gap events gain `class`; this proposal is a
  natural amendment there, not a separate file format.
- Dashboard: chain union-find generalizes to class grouping; judiciary section gains
  a per-class table (instances, recurrences, generator-fix status).
- Telemetry: one added map on the round line: `class_mass: {slug: mass}`.
- Related but separate: supersedes-edge grade-delta telemetry (accepted-delta
  blind-spot fix, same audit) — both want the merge to compute lineage-aware sums;
  implement together.

## Open questions for the debate

- Who arbitrates class definitions mid-run when red and blue disagree — the judge on
  the normal docket, or a dedicated ruling type?
- Recurrence N and the severity floor function: fixed policy or per-run parameter?
- Does class closure require generator-fix for ALL classes or only recurrence-tripped
  ones? (Proposed: only tripped ones — most classes are singletons and forcing
  generator hunts on every one-off is waste.)
- Retroactive classification of prior runs' records as a calibration corpus for the
  slug vocabulary?

---

# Addendum A: intra-round repair quality (from the round-4 provenance audit, 2026-07-17)

Baseline measured: round 3's 17 repairs produced 11 regression-successors (~50-65% of
repairs regress). Defect taxonomy from R4-1..R4-16: consistency (headline-vs-enum,
cross-section), false universals, arithmetic, repair composition, sibling enumeration,
repair-introduced overclaims, repair-minted degenerate cases, unverified cited anchors.
Diagnosis: shared literalism — red point-specifies, blue point-fixes, red point-verifies;
plus blue ships edits with no post-edit pass (rigorous at building, weak at rereading).

Levers, ranked by leverage-per-dollar (engine prompt changes to debate.js, land 0.9.0+):

1. **Acceptance check at mint.** required_fix gains VERIFICATION: the exact falsifiable
   check red will run at re-audit. Blue must run it pre-announcement and record the
   result. Red re-audit becomes spot-audit (cheaper), letter-vs-spirit gap dies at the
   instance level.
2. **Sequenced self-audit phase with per-repair manifest.** One manifest row per repair:
   change locus; figures recomputed; universal claims enumerated against cases;
   consistency sweep of touched paragraph + same-fact sections; boundary case of the fix
   ("what does this repair mint?"); interaction note when repairs share text. Capture
   audit verifies manifest completeness per closure (shape in-run, vacuity post-hoc).
3. **Repair minimalism + claim provenance.** Repairs change no more than the fix
   requires; new claims introduced by a repair are tagged verified-at-leaf / derived /
   asserted at write time; red targets `asserted` first.
4. **Class-sweep clauses (both sides).** Red: fix-spec states the class-closure rule or
   declares the enumeration open. Blue: repair enumerates siblings or declares the sweep
   boundary. Red: verifies against the class, not the instance list. (Prevents R4-2
   format-patch -o, R4-3 "includes" non-exhaustive list.)
5. **repair_regression_rate telemetry.** Round line gains closed_with_regression /
   closures + regression-successor mass; dashboard series; terminal report verdict input;
   sleeper-service consumable. This is the meter for "approaching zero."

Deliberately deferred: a dedicated proofreader seat (efficiency doctrine — cheapen
redundancy, never judgment; revisit only if the rate stalls >~15% with 1-3 in place).
Guard: the rate must be read jointly with red rigor signals (declined-to-mint counts,
spot-check density) — zero achieved by softer audit is a regression in disguise.

# Addendum B: relation of the three proposals

Gap classes (main), class-sweep clauses (A.4), and the repair-quality program (A.1-3,5)
are one thesis: instances are the unit of verification; classes are the unit of
resolution; repairs are the unit of durability. Classes handle recurrence ACROSS rounds;
the repair program handles defect injection WITHIN a round; telemetry makes both visible.

# Addendum C: blue's constitution — the missing telos (cross-run evidence, 2026-07-17)

Empirics under near-constant prompts (journals: retrospective r1-r5, efficiency r1-r4,
sleeper r1-r4 live):
- 13/13 red verdicts FAIL. No run has ever reached PASS; termination is always ceiling
  or human stop.
- Convergence curve is reproducible: mass deltas EFFIC −47.5/−28.3/−8.2, SLEEP
  −37.3/−26.5/−9.0 — plateau ≈40 by round 4 both runs; RETRO round 5 mass ROSE (27→28,
  6 minted > 5 closed): riding past the plateau goes backwards.
- Board turnover ≈100% every round in every run (blue repairs everything immediately).
- Later-round mints are 72–100% lineage-born (EFFIC r4: 10/10). The plateau is
  repair-churn equilibrium: rounds 4+ exist almost entirely to audit repair fallout.
- Dispute channel: 0 uses in ~13 rounds across 3 runs. Rulings: 76 carried, 1
  risk_accepted, 0 else.

Constitutional diagnosis (0.7.0 agents/, near-identical since run 3):
- RED has a terminal goal ("Adversary and gate-keeper... red decides when blue has met
  the bar. Never soft-pass"), gate ownership, and a learning loop ("record it in your
  project memory — the adversary learns").
- BLUE has duties, no telos: additive volume mandate ("union, never summary... grows
  every round"), a self-audit that is CITATION HYGIENE not correctness ("every claim
  footnoted, footnotes complete with access dates, confidence self-graded"),
  responsiveness ("address every gap"), and economics ("the pragmatist"). Blue has no
  goal in relation to red, no stake in the verdict, no memory of its own regression
  classes (it reads RED's memory; it accrues none). Red learns run-over-run; blue
  restarts from zero — predicting a constant-height plateau, which is what the data
  shows (40.8, 46).
- Precedent for amendment: blue's line 14 already hard-codes run 3's propagation lesson
  ("blue's dominant regression class") — constitutional learning has happened once, by
  hand. Institutionalize it.

Amendment (composes with Addendum A; engine 0.9.0+):
1. TERMINAL GOAL clause: blue's goal is a report TRUE AT THE LEAF — every claim ships
   only if it would survive the audit blue itself would run. Red's PASS is blue's win
   condition, reachable ONLY through durable repairs; winning by hiding, softening, or
   relocating material is defined as losing (anti-shell-game guard, ties to the conduct
   audit's dodge patterns).
2. Self-audit upgraded from citation hygiene to the correctness manifest (A.2);
   citations become one row of it.
3. BLUE MEMORY: a repair-regression pattern file symmetric to red's gap-pattern memory
   — "yesterday's regression class is today's free self-audit line." Symmetric learning
   loops; without it, red sharpens every run against a blue that never does.
4. repair_regression_rate (A.5) is blue's score, on the board and in the terminal
   report.

Falsifiable prediction: with 1-4, later-round lineage-born share and the plateau height
drop measurably; PASS becomes reachable. Baseline curves above are the control.
