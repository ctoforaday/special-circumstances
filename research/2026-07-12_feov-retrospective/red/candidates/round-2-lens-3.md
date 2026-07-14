# Round 2, Lens 3 — leaf-node citation verification (slice 3: §4 Friction ranking, §5 Open questions, Footnotes)

Full re-read of `blue/report.md` (704 lines) in context, plus `blue/CHANGELOG.md` (both rounds),
`debate.md` (full transcript), `red/findings.md` (round 1), `red/citation-ledger.md`, and the
primary corpus files `inputs/run1-friction.md`, `inputs/run2-friction.md`, `inputs/backlog.md`,
`blue/frontier.md`. Per protocol, claims already at HIGH confidence in the round-1 ledger and
untouched by round-1's CHANGELOG in my slice were not re-fetched; every claim below is either (a)
newly introduced/changed by blue's round-1 response inside §4/§5, or (b) a citation in my slice
not previously ledgered.

## New gaps

### R2-1 — OPEN — MEDIUM-HIGH — likelihood certain (direct comparison of citing text vs. cited source) x impact medium-high (inflates a likelihood grade and seeds a fabricated recurrence count into a run-4 trigger condition) x complexity low (correct the count, drop or replace the citation)

**Location:** §4, row 5 — *"Windows ENAMETOOLONG / long-command fragility | red (run 2 r1);
synthesizer (retrospective round 0); **red-merge (retrospective round 1, this round — third
occurrence, per debate.md's merge-seat friction)** | **3 runs/rounds, re-graded likelihood High
per R1-13** (was "2 runs")"*. Duplicated in §3, row 15 — *"confirmed recurred **3 times across 3
runs**, not 2 (this retrospective's own synthesis hit it a second time this round, at the
red-merge seat, per debate.md's round-1 merge-seat friction)"* — and in §5, item 10 — *"Does
ENAMETOOLONG recur a **fourth** time before the chunked-append helper is built?"*.

**Problem:** the cited source does not contain the claimed event. `debate.md`'s Round 1 "Merge-seat
friction" section (the only place in the transcript labeled "merge-seat friction") lists exactly
two items: (1) lossy PDF/HTML fetch depth, (2) a process misfit about a lens instance writing the
`### RED` transcript section before the merge ran. Neither mentions ENAMETOOLONG, a heredoc, a
shell-parse failure, or a command-length ceiling. The only two ENAMETOOLONG-class events actually
documented anywhere in the corpus are: (1) `inputs/run2-friction.md` line 4 (red-merge-r1, run 2,
"Bash heredoc hit ENAMETOOLONG... forcing me to split into ~6 append calls"), and (2) this
retrospective's own §0 addendum, which is itself dated to **round 0** (`blue/CHANGELOG.md`'s
Round 0 entry: *"LIVE EVIDENCE ADDENDUM (§0, §2.1, §3 #8, §4 rows 4–5): the Write of this very
report was refused by the write-block (third occurrence, third run), and the first
chunked-heredoc workaround failed on shell parsing"*) — not round 1. That is **two** documented
occurrences (run 2, and this retrospective's round 0), not three, and none of them is "the
red-merge seat, round 1." `inputs/run1-friction.md` (checked in full) contains zero ENAMETOOLONG
mentions, so "3 runs" is also unsupported on its own terms.

The likely mechanism: the ENAMETOOLONG narrative and the **write-block** narrative are structurally
adjacent throughout this report (both are the Tier-B "live-smoke-testable" pair, both use "third
occurrence" language elsewhere for the write-block's own, separately well-corroborated count — see
§4 row 4, R1-18, [^WriteBlock]) — round 1's fix appears to have transposed the write-block's
recurrence count onto ENAMETOOLONG's row without independently checking the ENAMETOOLONG-specific
source. This is exactly the citation-repair failure mode the ledger already tracks under
"repair-regression on citations" — a round-1 fix that touches a graded cell reads clean at a
glance but the citation backing the *new* number was never independently re-walked.

Downstream effect: §3 row 15's likelihood grade was re-graded Medium→High on this count; §5's new
item 10 frames the risk-accept as conditioned on "a fourth occurrence," which — if the real count
is 2, not 3 — should be phrased as a third occurrence, not a fourth. The disposition (risk-accept,
track for next occurrence) probably survives either way, but the grade and the open-question
framing are currently keyed to an uncorroborated fact.

**Required fix:** either (a) locate and cite the actual round-1 red-merge-seat ENAMETOOLONG
occurrence if one exists outside `debate.md`'s merge-seat friction section (e.g., an untranscribed
tool-call the merge seat made), or (b) correct the count to 2 documented occurrences (run 2,
retrospective round 0) across 2 runs, drop the "per debate.md's merge-seat friction" citation, and
renumber §5 item 10's trigger from "fourth" to "third."

**Corroboration confidence:** HIGH that the citation fails — direct read of the cited section
(`debate.md` lines 146–152) and both primary friction files confirms the claimed event is absent
from all three.

---

### R2-2 — OPEN — LOW-MEDIUM — likelihood certain (direct dating from the corpus's own CHANGELOG) x impact low (does not change the substantive "systemic across runs" conclusion) x complexity trivial (one-clause correction)

**Location:** §4, row 4 — *"3+ runs, 2+ filenames — systemic *across* runs despite low per-run
count, and now confirmed hitting a second role (red) in the same round it hit blue [L3, updated
round 1]"*. Related self-contradiction at the anchor claim's source, §0's "Live addendum" — *"red's
own round-1 audit pass hit the identical block writing `red/findings.md` **this same round**...
Two independent hits at two seats in **two consecutive rounds** is stronger evidence..."* (both
sentences in the same paragraph). Also propagates into §5, item 7 — *"Both write-block occurrences
observed **this round** (blue-synthesis, red-merge) hit ad hoc writes..."*.

**Problem:** chronology error. The blue-synthesis write-block hit is dated to **round 0** —
`blue/CHANGELOG.md`'s Round 0 entry explicitly logs it ("LIVE EVIDENCE ADDENDUM... the Write of
this very report was refused by the write-block"), and the report itself labels it "synthesizer
(retrospective round 0)" in §4 row 5 one row below the claim being challenged here. Red's
write-block hit on `red/findings.md` is dated to **round 1** — `debate.md`'s Round 1 RED section:
*"the report-file write-block... fired again on this red pass's own attempt to write
`red/findings.md` mid-task"*. These are two different rounds (0 and 1), not "the same round." The
source paragraph in §0 says both things in adjacent sentences — "this same round" then "two
consecutive rounds" — which is internally contradictory on its face; §4 row 4 and §5 item 7 both
inherited the "this same round" framing rather than the (correct) "two consecutive rounds" framing
from the same paragraph.

**Required fix:** change "in the same round it hit blue" (§4 row 4) and "observed this round"
(§5 item 7) to "one round after it hit blue" / "across rounds 0 and 1" respectively; fix the
internal contradiction at the source in §0 by dropping or correcting "this same round."

**Corroboration confidence:** HIGH — both events are independently dated by the report's own
CHANGELOG and debate transcript; the contradiction is legible within a single paragraph.

## Verified clean this round (leaf-node re-check against primary sources, my slice)

- §4 row 6 correction (R1-8): *"run 2, round 1 only"* — `inputs/run2-friction.md` contains
  exactly one live-source-drift entry (line 6, red-merge-r1); no round-2-scoped instance found.
  **HIGH.**
- §4 footnotes [^Run2Friction]/[^FrictionCount] correction (R1-7): *"21 entries"* — direct count
  of `inputs/run2-friction.md` bullet lines (3–23) = 21. **HIGH.**
- §4 rows 7–9 role/round attributions ("red only", run 2 r3/r4 respectively) — line-by-line match
  against `inputs/run2-friction.md` lines 17, 21, 22 (trajectory-extractor opacity, Auto Dream
  sandbox, Springer auth-wall, all red-merge-r3/r4). **HIGH.**
- §4 row 10 long-tail singletons ("%TMP% clobbering in the doctor bootstrap," "missing workflow
  progress heartbeats") — traced to `inputs/backlog.md` line 7 and `ideas/backlog.md` line 29
  respectively (not `run2-friction.md`, correctly uncited to it). **HIGH.**
- §4 shape-verdict's frontier quote — *"the write-block/ENAMETOOLONG/preflight-guard complaints
  cluster in round 1 only (already fixed)"* — verbatim match, `blue/frontier.md` lines 45–46.
  **HIGH.**
- §5 items 8, 9 cross-references to §3 rows 19 and 16b — content matches the referenced rows
  verbatim in substance. **HIGH** (internal consistency, not an external citation).
- [^SubagentWriteBug] (R1-9's softened form, as it now reads in the footnote) — consistent with
  round-1 ledger grading (MEDIUM); no change this round, not re-fetched per the freshness rule
  (CHANGELOG shows no further edit to this footnote in round 1 beyond the round-0→round-1 fix
  already ledgered).

## Note on scope

This slice (§4, §5, footnotes) inherits two citations from §3 (row 15) that duplicate the R2-1
finding verbatim — flagged here since the underlying miscitation is the same fact wearing two
section numbers, and my assigned rows (§4 row 5, §5 item 10) are the ones that actually carry it
forward into a grading decision and an open-question trigger condition.
