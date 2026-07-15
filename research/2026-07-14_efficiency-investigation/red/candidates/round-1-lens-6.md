# Red candidate pass — round 1, lens 6: dark-side and risk

Auditor: red lens 6 (failure modes, likelihood × impact × complexity grading, security and
tradeoff blindspots). Audit surface: FULL `blue/report.md` (914 lines, read whole in two
windowed pages — 25k Read cap; see friction). Engine claims leaf-verified against
`plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js` read in full at
the working tree, with pin equivalence to `5396952` re-confirmed first-hand this pass
(`git diff --stat 5396952 -- ideas/backlog.md plugins/frank-exchange-of-views/` empty —
guarding the self-referential-repo-drift class; repo HEAD now `b162e50`, which touches only
this run's own skeleton).

Overall lens verdict input: FAIL (no finding below is report-sinking alone; F1 and F2 are
gate-relevant and must be answered before PASS).

---

## L6-F1 — The efficiency analysis prices run 4 against a dead baseline: shipped auto-docket guarantees a judge-seat cost line (and per-round board disposition) that no section models

**Location:** §6.1 "Where the money actually is" — "From cost.md at the pin: red-merge $57.54
(38%) > red-lens $49.48 (33%) > blue-respond $18.21 (12%) > blue setup/synthesis + assembly
(one-time)." Also §3.1 — "any gap that stays open across two rounds now auto-dockets — a
dispute red re-rejects already reaches the judge under shipped mechanics." Also §1.2 — "A
judge disposing the round-3 residual board does not audit; disposition produces no R4-1."

**Problem.** Blue's §3.1 characterization of the shipped detector is *correct* — I verified it
at the source, and it is stronger than blue uses it for. `debate.js` lines 244–245: `contested
= redEnv.gaps.filter(g => allPriorGapIds.has(g.id) || (g.supersedes || []).some(id =>
allPriorGapIds.has(id)))`. `redEnv.gaps` is the full open board (the blue-respond dispatch at
line 261–263 derives "every open gap" from exactly this array), so from round 2 onward **every
persisting open gap dockets to the judge, every round**. Three consequences the report never
prices or reasons about:

1. **A new recurring cost line at the judgment tier.** The judge dispatch (line 249) reads
   `debate.md` AND `red/findings.md` **in full** — the same cumulative-context TURNS × CONTEXT
   driver §4.2 diagnoses at the merge seat, at the same 5×/12.5× rates. On a run-3-shaped
   board (11/10/5/6 persisting opens after rounds 2–5) that is ~4 judge dispatches per run,
   each of the same context class as a $10–13 merge. §6.1's money map, the lever rankings, and
   every savings estimate (§2.4's "~$12–18/run", §4.2's "$5–15/run") are computed against
   run 3's **zero-judge** cost distribution, which shipped PR #15 mechanics have already made
   unreachable. The error term is plausibly the size of the savings being ranked.
2. **Carried gaps re-docket forever.** Line 253: only `closed | rebuttal_sustained |
   risk_accepted` enter `adjudicated`; a `carried` ruling leaves the gap in the board, so it
   re-dockets and is re-ruled **every subsequent round** until closed. Repeated
   judgment-tier spend, plus a fatigue path on red's gate: each re-docket is a fresh
   opportunity for the ruling to drift carried → risk_accepted, which removes the gap from
   red's verdict permanently (`adjudicated` is append-only; no re-open path except a fresh
   supersedes successor).
3. **§1's rejection stands on a baseline the shipped code already moved.** Blue rejects
   severity-floor termination partly because "a judge disposing the round-3 residual board
   does not audit" and rejects lane 2's judge-routing variant as unproven — while the shipped
   engine already routes *more* than lane 2's variant ever would (the whole persisting board,
   unconditionally, every round — no ≤-MEDIUM arming condition at all). The §1.5 carried
   option's open interaction (open question 8, PASS-path vs degenerate-FAIL guard) is not a
   future design question; a whole-board disposition by the judge is reachable in run 4 under
   shipped mechanics, and the state machine for "judge closed everything, red's next verdict
   is PASS-by-exclusion" is unexamined. Symptom that blue has not internalized this: §8 open
   question 4 asks whether "the lineage docket arm[s] in its first live trial" as if arming
   were uncertain — under the code it is near-certain in any multi-round run with one
   persisting gap.

**Not claimed:** that the mechanic is wrong (it shipped deliberately, winnow list bars
re-litigating it). Claimed: an *efficiency investigation* whose deliverable is a
cost-priority ranking must model the cost structure the next run will actually have.

**Grading:** likelihood HIGH (mechanical certainty given any persisting gap; run-3-shaped
boards persist heavily) × impact MEDIUM-HIGH (cost-model error of the same order as the
savings being ranked; plus the ungraded carried-gap re-docket/fatigue path on red's gate) ×
complexity LOW-MEDIUM (add a projected judge-seat line to §6.1 keyed to expected docket
rounds; state the §1 interaction honestly; flag carried-gap re-docket dedupe as a design
question, not necessarily a change).

**Required fix:** (a) add the judge seat to §6.1's money map as a projected run-4 line with
the arming analysis above; (b) correct §1.2/§1.5 and open question 4 to state that the docket
arms near-certainly and that per-round whole-board judge disposition is *already live*, then
re-derive whether the carried minority option adds anything the shipped mechanics don't; (c)
grade the carried-gap re-docket loop (cost + gate-erosion) or cite where it was
risk-accepted.

**Corroboration:** HIGH — code read first-hand at verified pin; report quotes cross-checked
in place.

---

## L6-F2 — The grade-dispute channel (the interlock that "makes actuation safe") audits only its rejected branch, and red's PASS is an escape hatch that moots pending disputes

**Location:** §3.3 — "the script holds rejected disputes one round and adds re-disputed ids
to `contested` — set arithmetic mirroring the existing lineage filter"; §3.5 — "BINDING
interlock (all three lanes converge on this): the channel is a mandatory companion of any
future lever-1/2 actuation."

**Problem.** Two schema-legal control-flow holes plus one incentive asymmetry, in the very
mechanism §3.5 and §6.2 present as the safety companion for grade-actuated spend:

1. **Accepted disputes are unaudited by design.** The proposed flow dockets a dispute only
   after red REJECTS it and blue re-disputes. A dispute red ACCEPTS changes a grade in
   `red/findings.md` with no third-party record, no docket, no check. Under the future
   actuation this interlock exists for, the cheap collusion path is exactly there: blue
   disputes grades downward (lower mass → fewer lenses auditing blue next round), red accepts
   because accepting is cheaper than arguing (the report itself documents fix-because-cheaper
   behavior as the channel's motivating case). Mass deflates; the throttle starves the
   adversary; the judge never sees it. §2.3 names red's self-grading incentive loop; the
   channel adds blue's symmetric deflation lever and the report grades only the first.
2. **PASS moots pending disputes.** In the shipped loop, `if (redEnv.verdict === 'PASS')
   break` (line 236) executes BEFORE the contested-docket block (line 244). Unless the
   channel's filter is explicitly ordered ahead of the PASS break, a red merge carrying
   rejected-or-unaddressed disputes can end the run by issuing PASS — the dispute dies
   unadjudicated and the contested grade ships in the permanent record §3.2 says is the
   channel's whole point (cross-run docket assembly). Same terminal hole at maxRounds: a
   final-round dispute is never held anywhere. This is lane 1's §3.4 observation
   ("final-round grades are never adversarially audited") applying to the channel itself —
   the channel as specified does not fix the terminal case it is most needed for, and the
   report does not notice the recursion.
3. **Unbounded dispute traffic is a cost lever pointed at red.** `grade_disputes` has no
   cap; each disputed gap_id × dimension obligates a `dispute_responses` entry on pain of
   auto-docket. Blue can force red-merge turns or mass judge dispatch at near-zero envelope
   cost. Low likelihood with honest seats; it becomes the rational play the moment grades
   control blue's audit burden — i.e., precisely under actuation.

**Grading:** likelihood LOW this run (no actuation, zero observed dispute traffic — blue's
own §3.1 finding) rising to MEDIUM-HIGH under the actuation the interlock is mandatory for ×
impact MEDIUM-HIGH (blue's own framing: "wrong grades... start being budget errors"; a
permanent-record integrity mechanism with an unaudited write path fails its one job) ×
complexity LOW (three sentences of design: accepted-dispute grade deltas emitted in the
§1.5 board-profile log line and eligible for the §4.5-style spot-check floor; dispute
disposition ordered before the PASS break, or PASS-with-pending-disputes auto-dockets them;
a per-round dispute cap or batch-docket rule).

**Required fix:** amend §3.3 with the three clauses above (or argue any of them
risk-accepted with the incentive analysis stated). The interlock claim in §3.5/§6.2 should
not survive unqualified while the accepted branch is dark.

**Corroboration:** HIGH on the control-flow facts (code + report's own §3.3 text);
MEDIUM on the incentive projection (no corpus instance — structural argument, same
evidentiary tier as blue's own §2.3 incentive-loop claim, which blue treated as load-bearing).

---

## L6-F3 — §4.5's safety conditions are themselves the unenforced-prompt-MUST failure class the report elsewhere treats as proven

**Location:** §4.5 — "**Archive spot-check floor** — N sampled closed cases re-verified per
round, never zero"; and conditions 2 ("a named MUST for archive verification") and 4
("reopen + drift triggers").

**Problem.** The report's own §3.3 inherits "the R5-5/R3-2 unenforced-optional-field lesson
(a schema'd field under prompt instruction alone goes silently unset; three rounds unnoticed
in run 3)" — and applies it to lever 3a's default-to-docket. Conditions 2, 4, and 5 of the
ratified lever 4a are all prompt-level MUSTs with no observable, no schema field, and no
structural check: nothing in the engine or envelopes can detect that the spot-check floor
quietly hit zero, that a lineage claim was verified from memory instead of by archive read,
or that a reopen trigger was skipped. The failure is silent by construction and its
consequence is precisely the R5-1 stale-lineage class the conditions exist to keep dead.
Inconsistent application of the report's own named failure class: schema enforcement for 3a,
hortatory MUSTs for 4a.

**Grading:** likelihood MEDIUM (in-corpus base rate: R3-2's schema'd-but-unset field ran
three rounds unnoticed) × impact MEDIUM-HIGH (silent return of the stale-lineage/archive-rot
class under the new sharded structure, undetectable until a run-3-style catch) × complexity
LOW (when the sharding PR ships: RED_ENVELOPE gains e.g. `archive_spot_checks: [gap_ids]`
required non-empty from round 2, mirroring the shipped lineage-throw pattern; closure-index
line count reconciled against archive block count — one grep).

**Required fix:** add a seventh condition making conditions 2/4/5 observable (schema or
structural check), or state explicitly why prompt-level MUSTs are acceptable here when §3.3
rejects them for the dispute channel.

**Corroboration:** HIGH (report-internal inconsistency; both halves quoted from the audit
surface).

---

## L6-F4 — `found_by` is self-reported by the seat whose budget it would later control, and the report parks the problem in an open question while a mechanical check sits in the run directory

**Location:** §2.5 — "Record **lens overlap per merged gap**... a `found_by` envelope field
(e.g. `['L1','L2','L4']`) — red-merge already states convergence in prose; this is the
capture-recapture input, mechanical to collect"; §8 open question 3 — "Can red-merge
attribute lens overlap honestly... without inflating convergence."

**Problem.** The instrumentation is ratified 3/3 with no verifiability condition, while
lever 4a's ratification carries six. `found_by` is written by red-merge — the seat whose
lens budget a future capture-recapture throttle would set from this very field. Two failure
directions, both invisible in the envelope: under-reported overlap inflates the
remaining-defect estimate (keeps red's budget high — red's incentive under actuation);
clustering judgment ("are these two lens findings the same gap?") sets the denominator and is
exactly where honest drift lives. The report itself proves the check is cheap: the lens
candidate files are preserved on disk (`red/candidates/round-N-lens-*.md` — §2.3's R5-5
singleton claim was verified against them), so `found_by` is grep-auditable per gap. Parking
honesty in an open question instead of a ratification condition means runs 4–5 generate the
actuation evidence base with an unaudited instrument.

**Grading:** likelihood MEDIUM (attribution is judgment under incentive, not arithmetic) ×
impact MEDIUM (corrupts the evidence that decides future actuation — the report's own
evidence-before-mechanism standard) × complexity TRIVIAL (one ratification condition:
`found_by` values are auditable against the preserved lens files, and the §4.5-condition-5
spot-check floor samples them).

**Required fix:** promote open question 3's honesty concern to a named condition on the
ratified instrumentation, with the lens-file audit as the mechanism.

**Corroboration:** HIGH (report-internal; lens files' existence verified in the run-3 corpus
citations blue itself uses).

---

## L6-F5 — The degenerate collator's one sentence has an unpinned output path: written into `candidates/`, it double-counts every finding for downstream readers and double-indexes the corpus

**Location:** §4.6 — "one added prompt sentence — first action, `cat
red/candidates/round-N-lens-*.md > round-N-all.md`, then read the single file."

**Problem.** As written the redirect target is cwd-relative, and agent cwd resets between
bash calls in this harness — the file lands wherever the seat happens to be, or (if the
natural fix is applied and it is written into `red/candidates/`) it becomes a sixth file in
the directory every downstream consumer globs or lists: the merge prompt's "read the
round-N lens passes in `red/candidates/`", any future `found_by` audit (L6-F4), any
mechanical convergence count, and report-assembly unions. Every finding then appears twice;
a mechanical convergence counter reads every singleton as 2-of-6 — and convergence counts
feed corroboration grades (§2.3: "four of five lenses converged" is part of R4-1's HIGH).
Additionally the `sc-recall-index` hook fires on every markdown write, so the concatenation
duplicates all lens content in the qmd index — mode-1 retrieval returns doubled hits.
Untested filename risk is nonzero too: the write-block guard is filename-keyed and
demonstrably broader than `findings.md` (it refused `blue/report.md` at blue's own synthesis
seat THIS round, per §4.3(c)).

**Grading:** likelihood LOW-MEDIUM (agent judgment usually distinguishes the file; mechanical
consumers are exactly the ones that won't) × impact LOW-MEDIUM (inflated convergence →
inflated corroboration grades; doubled retrieval; wasted merge turns) × complexity TRIVIAL
(pin the target to the session scratchpad or a non-`candidates/`, non-indexed path with an
excluded name; one clause in the same sentence).

**Required fix:** amend the §4.6 batching sentence to name an absolute output path outside
`red/candidates/` and outside the recall index's watch surface.

**Corroboration:** HIGH on the glob/consumer facts (merge prompt read at source, line 216);
MEDIUM on the qmd double-index claim (hook behavior per protocol text, not exercised).

---

## L6-F6 — Ratified instrumentation names a writer that cannot exist and an unversioned metric

**Location:** §2.5 — "Compute and `log()` sum(L × I) per merge — grades are already
machine-readable in `redEnv`... one arithmetic line, zero new seats; emit into `cost.md` per
round."

**Problem.** Two specification defects in the one lever every lane ratified. (a) No
component can "emit into cost.md per round": the script has no filesystem access by design
(`debate.js` lines 32–34, the report's own [^EngineSource]), and `cost.md` is produced
post-run by `scripts/cost-audit.mjs` from token metering, which does not see gap grades. As
specified, the destination is unreachable — policy without a mechanism. The honest path is
the `log()` line landing in `trajectories/journal.jsonl` and `cost-audit.mjs` (or the
retrospective) consuming it. (b) The mass series will be logged under an undecided mapping —
§2.5's own flagged design detail and open question 6 ("does `realized` count in open-gap
mass at all") — so if the mapping is decided differently later, runs 4–5's series is
incomparable with successors, poisoning exactly the evidence the deferred actuation decision
needs. Decide the mapping now or version-stamp the log line.

**Grading:** likelihood CERTAIN-as-specified for (a) (the sentence cannot execute as
written), MEDIUM for (b) × impact LOW (instrumentation only; correctable in the PR) ×
complexity TRIVIAL (name journal.jsonl as the sink; pin or version the enum→numeric mapping).

**Required fix:** reword §2.5 item 1 with an executable sink and a pinned/versioned mapping.

**Corroboration:** HIGH (code + report text side by side).

---

## L6-F7 — The carried minority floor arms on a single zero-mint round while its own cited literature requires double confirmation, and cannot distinguish discovery decay from a degraded red round

**Location:** §1.5 — "arms only when BOTH (a) every unadjudicated open gap ≤ MEDIUM with
low/trivial fix cost AND (b) the round just completed minted zero new gaps above the floor
(the discovery-decay arm, Good-Turing-shaped)."

**Problem.** §1.3 cites the adaptive-stopping criterion as "distributional change below ε
for **2 consecutive rounds — a double confirmation**"; the carried variant arms on ONE
zero-mint round. A single such round is also exactly what a degraded red round produces
(lens under-delivery, the run-2 null-return class, a cheap-model round per the row-16b
tradeoff) — the arming condition cannot distinguish "nothing left to find" from "the finder
failed this round." Impact is bounded because the variant routes to the judge and never
terminates, and because it is carried, not built — but carried designs get built later from
this text, and the report registers a prediction for red against this very variant.

**Grading:** likelihood LOW (variant unbuilt; degraded-round base rate ~1 instance in 3
runs) × impact MEDIUM if ever built (judge disposition triggered by adversary failure is the
severity floor's §1.2 defect wearing a new arming condition) × complexity TRIVIAL (arm on
two consecutive qualifying rounds + a red-health sanity term).

**Required fix:** carry the variant with the double-confirmation and red-health conditions
attached, so the registered design is internally consistent with §1.3's cited criterion.

**Corroboration:** HIGH (report-internal inconsistency, both halves quoted).

---

## L6-F8 — Condition-6 shard names are untested against a write guard the report itself shows is broader than one filename, and two engine prompts hardcode the path being renamed

**Location:** §4.5 condition 6 — "non-report-semantic names (e.g. `red/ledger.md`,
`red/archive.md`)"; §4.3(c) — "blue's own synthesis Write of `blue/report.md` was refused by
a filename-keyed guard at the synthesis seat."

**Problem.** The write-block guard's key set is unenumerated and demonstrably wider than
`findings.md` — it refused `report.md` this round. Proposing specific shard names as safe
without a preflight test risks a mid-merge write failure (turn waste, detour) in the first
sharded run. Smaller but real: `debate.js` hardcodes `${runDir}/red/findings.md` at both the
red-merge prompt (line 216) and the judge prompt (line 249); renaming the files without the
engine change strands the judge's full read — the report's condition list never enumerates
the engine-side prompt edits the rename implies.

**Grading:** likelihood LOW-MEDIUM (Edit-on-skeleton-born files worked all of run 3 — the
main path is proven; Write/`cat >` fallbacks and fresh-name creation are where the guard
bites) × impact LOW (known detour exists; cost is turns and confusion, not data loss) ×
complexity TRIVIAL (preflight-test candidate names when the skeleton PR lands; enumerate the
two prompt-site edits in the sharding PR's scope).

**Required fix:** add name-preflight and the two engine prompt sites to condition 6.

**Corroboration:** HIGH (guard behavior from the report's own live instance + run-3 friction
#4; prompt sites read at source).

---

## L6-F9 (advisory) — The lever-5 gate treats run-4 non-detection as positive evidence with an instrument of unmeasured sensitivity

**Location:** §5.5 — "contingent on run 4's propagation-clause record showing zero
unpropagated-site regressions."

**Problem.** "Zero regressions observed" and "zero regressions" differ by the detector's
sensitivity, which §5.5 itself concedes is unmeasured ("the floor's catch probability is
unmeasured"). A run-4 red that happens not to run the propagation greps produces a clean
record that ratifies scoping — the gate can be satisfied by detector under-performance. The
report is honest about the post-hoc arms; the missing clause is that the gate should require
positive evidence the detector RAN (propagation-grep arm exercised and logged per accepted
correction), not merely absence of findings.

**Grading:** likelihood LOW-MEDIUM × impact MEDIUM (ratifying an unsafe selection rule on a
false-clean record — §5.3's own safe-RTS framing) × complexity TRIVIAL (one sentence in the
gate: the run-4 record must show the greps were run, not just that nothing was found).

**Required fix:** amend the §5.5 gate with the detection-effort requirement.

**Corroboration:** HIGH (report-internal; the gate text and the unmeasured-sensitivity
concession are three paragraphs apart).

---

## Checks run that produced no finding (for the merge's convergence accounting)

- §2.3's `citationPasses` quote matches source (line 198; report omits the `|| 20` default —
  immaterial).
- §3.1's zero-LEAD structural grep, §1.1's never-fires table, and §2.1's mass series were
  not re-derived by this lens (citation/logic lens territory); no dark-side defect turns on
  them.
- Doctrine check §6.3 holds against every finding above: none of F1–F9 asks blue to cheapen
  judgment or the adversary; F1/F2 ask for pricing and audit of judgment-seat traffic, which
  strengthens the gate.
- Reflexivity: the report applies its context-degradation citations to merge and lens seats
  (§4.2, §5.3) and logged its own live write-block instance — no reflexivity blindspot beyond
  the judge seat, which F1 covers.

## Friction (lens seat)

- 25k-token Read cap vs `blue/report.md` (914 lines, ~32k tokens): full re-read required two
  windowed reads. Same class as run-3 friction #15, now recurring at a lens seat on a
  *round-1-sized* report — it will worsen every round. Wanted: a Read mode (or sanctioned
  pattern) for "whole file, I accept the token cost" at audit seats where the protocol's
  full-read clause outranks the saving.
