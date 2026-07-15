# Round 4 — lens 6: dark-side and risk (failure modes, likelihood × impact × complexity, security and tradeoff blindspots)

Audit surface: FULL re-read of `blue/report.md` (1535 lines, ~56k tokens, three windowed
reads — the 25k Read cap forced windowing again; friction noted below). CHANGELOG used for
navigation only. Focus: the round-3 repairs as new mechanism claims — each audited for its
own failure modes, executor, timing, and incentive interactions, per the lead's round-4 bar
("no design clause ships without naming who executes it and confirming the named channel can
physically carry it") — plus the corrected-dollar family for un-recomputed siblings.

Leaf verifications performed at this seat (positive results first — ledger-worthy):

- **The R3-1 instrument re-run FIRST-HAND at this seat**: extracted the pinned run-3 tarball
  (46 members) to the session scratchpad and ran
  `node trajectories/decompose-merge.mjs <extracted-dir>`. Every printed figure reproduces
  exactly: raw shares (r1 35.7/0.1/46.5/4.0; r2 36.3/8.9/28.5/6.9; r3 43.0/16.7/21.0/10.3;
  r4 20.1/31.7/33.6/6.0; r5 46.0/28.8/18.9/2.3), per-round totals 172.7/245.8/249.2/188.2/
  316.6KB, strict findings dollar series $0.00/0.26/0.53/0.89/1.16 Σ=$2.84, "other" bucket
  13.6/19.5/9.0/8.6/4.0%, round-5 61 assistant turns, blue ingest 145.7KB, lens-candidate
  ingest 52.4–80.3KB/round. The script applies bytes/4 at the pricing step (l.75), matching
  the documented method. **R3-1's substance repair verifies at the leaf.** HIGH.
- **Archive fraction reproduced**: the quoted awk run against pinned findings.md returns
  76,356 / 105,223 = 72.57% ≈ 72.6%, and l.340 is verbatim "## Verdict (round 4): FAIL —
  superseded by round 5, preserved". HIGH. (Nit, checked and not raised: gawk `length()`
  counts UTF-8 characters, not bytes — awk sum 105,223 vs `wc -c` 106,772; numerator and
  denominator are equally affected and the fraction is robust; "body bytes" is a hair loose.)
- `debate.js` read first-hand at the pin-equal tree: GRADE enum l.60
  (`low|low-medium|medium|medium-high|high|certain|realized|trivial`); `gaps[].likelihood`
  and `impact` typed GRADE (l.88); judge resolution enum incl. `carried`; the LENS dispatch
  prompt names `blue/report.md` + CHANGELOG (+ citation-ledger for citation instances) —
  **debate.md is NOT on the lens read surface**; blue-respond reads the latest `### RED` and
  `### LEAD`; the judge reads debate.md in full only when dispatched. Load-bearing for
  L6-F3, L6-F5.
- `.gitignore` re-read live (repo-root, OUTSIDE every pinned path — drift-checked per the
  self-referential-repo rule): still exactly one trajectories entry
  (`**/trajectories/agent-transcripts.tar.gz`); board-telemetry premise still sound.
- Recomputed clean: $2–4 = 72.6% × $3–6 ✓; proportional ceiling ≈$6 ✓ (findings shares ×
  ~58%-cache-read merge dollars ≈ $6.2); §2.1 per-gap means 4.9/5.9/4.4/6.0/5.2 ✓; ~21 =
  10+5+6 ✓; "52–80KB/round" ✓ from the re-run.

## Findings

### L6-F1 — the instrument is "committed" in three artifacts and tracked in none

**Severity: LOW-MEDIUM — likelihood medium (the claim is false at audit time; a run-end
whole-dir sweep is plausible precedent but no named mechanism covers a `.mjs`) × impact
medium (the R3-1 centerpiece repair's preservation half; §6.2's vacuity tier requires
git-tracked artifacts, and the reproduction promise dies with the working tree in the
interim) × complexity trivial. Corroboration: HIGH (first-hand: `git status --porcelain`
shows `?? .../trajectories/`; `git ls-files` for the run dir returns no trajectories entry;
the run dir's only commits are the two 06:02 staging commits).**
**Lineage note for the merge: successor-candidate to R3-1 — the defect is the repair's own
status word.**

**Location:** §4.2 — *"The instrument is now **committed as
`trajectories/decompose-merge.mjs`** in this run's directory"* — propagated to
[^MergeDecomposition] ("the method is now committed"), CHANGELOG Round 3 ("REBUILT and
COMMITTED"), and debate.md round-3 BLUE ("The parser is now committed").

**Problem:** the file exists in the working tree, executable and correct (re-run verified
above), but it is untracked: not in the index, not in the object store, in a directory git
has never seen. "Committed" is a past-tense work-done claim that is false in the only sense
R3-1(b) cared about — R3-1's stated harm was precisely that the instrument's artifacts were
"exactly the ones that are not git-tracked," and §6.2 places vacuity checking "over
git-tracked artifacts." The whole R3-1 correction record polices derivation-status words
("measured" vs "modeled"); "committed" vs "written to the working tree" is the same word
class, in the same correction record. Mitigants, stated honestly: run-3 precedent shows a
run-close sweep tracked `candidates/` and `trajectories/journal.jsonl`, so eventual tracking
is plausible — but the run-record capture step (research.md step 7) enumerates only
journal.jsonl + tarball + cost.md; nothing names an instrument-class file, and the tarball
input is likewise still delete-able until close. The interim window is real and the manifest
silence means the sweep is convention, not mechanism.

**Required fix:** commit the instrument now (it is not gitignored; one `git add` at any
seat/operator with commit rights), or restate the status word honestly at all three sites
("written into the run directory; git-tracked at run-record capture") AND name
instrument-class files (`trajectories/*.mjs`) in the capture manifest so preservation is a
mechanism, not a hope.

### L6-F2 — §6.1's "comparable to batching" compares $/run against $/round, and the batching figure is the un-recomputed sibling of the 4×-corrected series

**Severity: LOW-MEDIUM — likelihood certain (textual units; recompute run first-hand) ×
impact low-medium (the money map's #1-vs-#2 ordering rationale; no disposition flips — both
items are ratified regardless) × complexity trivial-to-low. Corroboration: HIGH (both
figures quoted from the report; strict-conversion ceiling recomputed at this seat from the
committed instrument).**
**Lineage note for the merge: successor-candidate to R3-1 — the sibling-halo residue of the
same pricing correction.**

**Location:** §6.1 item 1 — *"sharding-addressable ≈$2–4/run at run-3 scale, comparable to
item 2's batching saving rather than dollar-dominant"* — vs §4.6 item 2 — *"collapsing 6–8
read turns ... saves roughly $1–2/round at round-5 merge rates."*

**Problem:** two defects. (a) **Unit mismatch:** $2–4/RUN is compared to $1–2/ROUND;
normalized over run 3's five merge rounds, batching's printed figure is $5–10/run — which
would not be "comparable," it would dollar-dominate sharding 2×. (b) **The batching figure
was never re-derived after R3-1's conversion correction.** It dates from the round-1/round-2
text and its derivation is unstated; under the strict (corrected) conversion, this seat's
re-run prices the ENTIRE lens-candidate line at $0.61–0.88/round cache-weighted — the
claimed $1–2/round saving exceeds the whole cost of carrying the material it batches, which
is only possible if the figure inherits the same bytes-as-tokens ≈4× overstatement the
committed parser just corrected in its sibling. Honest recompute (turn-collapse ≈ removed
turns × context re-billing, strict conversion) lands near $0.3–0.5/round ≈ $1.5–2.5/run —
which would RESTORE the printed "comparable" and arguably sharding's dollar edge. Either
way, the printed comparison sets a corrected figure against an uncorrected one in different
units, in the sentence that ranks the run's top two actionable items — the exact
sibling-halo class R3-1 documented (one figure in an artifact corrected, its
same-method sibling left standing).

**Required fix:** re-derive the batching saving from the committed instrument's own output
under the strict conversion, state it per-run, and restate the §6.1 comparison in one unit
("batching ≈$X/run vs sharding ≈$2–4/run"); if the re-derivation is deferred, mark the
batching figure with the same modeled-not-measured status the R3-1 correction imposed on its
sibling.

### L6-F3 — condition 6's preflight now contradicts "skeleton-born" and fires after the point of no return

**Severity: LOW-MEDIUM — likelihood medium (ratified spec, would be built as written; the
two clauses of the same condition name different creators for the same files) × impact
low-medium (the guard can only fail AFTER the sharding PR ships — first sharded run,
round-1 merge, prompts at ll.216/249 already renamed and the judge's full read stranded;
manual recovery mid-run) × complexity trivial. Corroboration: HIGH (report-internal, both
clauses quoted; every corpus write-block fact already ledgered). Vector honesty, stated
plainly: the "opening act" example is red's own R3-11 required-fix phrasing, repeated by the
lead's ruling — red's phrasing is the vector, third instance of the class this run; the hole
is still a hole.**
**Lineage note for the merge: successor-candidate to R3-11.**

**Location:** §4.5 condition 6 — *"both files pre-created in the blackboard skeleton (PR #14
pattern)"* — vs the same condition's round-3 addition — *"Concretely: the first sharded
run's red-merge writes both skeleton shards as its opening act; alternatively, verify the
guard's seat-independence first-hand and cite the verification in the PR."*

**Problem:** two defects in the repair's own text. (a) **Internal contradiction:** the
condition's headline says the shards are pre-created by the skeleton step (a
lead/orchestrator act — the seat class R3-11 ruled inadequate as preflight); the round-3
addition says red-merge creates them as its opening act. Both cannot be the creator. If the
skeleton pre-creates, the red-merge "opening act" Write is either skipped (the vacuous
preflight recreated — red-merge only ever Edits/cat-appends, so the guard's Write behavior
at that seat class is never exercised) or an overwrite that quietly abandons the
skeleton-born provenance §4.3(c)'s write-path safety case rests on. (b) **Timing:** as
scheduled, the preflight can only fail INSIDE the first sharded run — after the run-5 PR has
shipped the renamed prompts and retired the monolith. A guard that can only fire after the
decision it gates is committed is a smoke alarm wired to the ashes; no fallback is stated
for a blocked opening-act Write with the judge prompt already pointing at the new names.

**Required fix:** two sentences — (i) reconcile the creator: name which seat creates the
shard files and update "skeleton-born" accordingly; (ii) reschedule the preflight BEFORE the
PR ships: a red-merge-class seat in a LIVE run test-writes the proposed shard names as a
closing act of run 4 (this run's own red-merge is exactly the right seat class and costs two
Write calls), and the PR cites that preflight — the "verify seat-independence first-hand"
alternative branch already has the right shape and should be promoted from alternative to
default if the closing-act preflight is not run.

### L6-F4 — the pinned mapping omits `trivial`, a schema-legal likelihood/impact value

**Severity: LOW — likelihood low (red's conventions emit `trivial` only for complexity, but
the schema permits it for likelihood and impact) × impact low-medium (one improvised value
mid-series is exactly the comparability break the pin exists to prevent — the R3-8 harm
restated) × complexity trivial. Corroboration: HIGH (GRADE enum read first-hand at
debate.js l.60; `likelihood: GRADE, impact: GRADE` at l.88).**
**Lineage note for the merge: successor-candidate to R3-8 — the pin is under-inclusive
against the enum it maps.**

**Location:** §8 Q6 — *"The pinned mapping for runs 4–5's telemetry series: low=1,
low-medium=1.5, medium=2, medium-high=2.5, high=3, certain=3.5, realized=excluded —
version-stamped into each logged line"* — and §2.5 item 1's mapping-stability condition.

**Problem:** the engine's GRADE enum has eight values; the pinned mapping covers seven.
`trivial` is schema-legal for `likelihood` and `impact` (both typed GRADE), and a gap
graded, e.g., `trivial × high` gives red-merge no mapped value — the seat improvises or
skips, and either choice is an unversioned convention change inside a series whose whole
design premise (R3-8, accepted round 3) is that mid-series convention changes poison the
actuation evidence. A mapping "pinned before the first logged round" that does not cover
its input domain is pinned in name.

**Required fix:** one clause in the Q6 pin: either `trivial=0.5` or "trivial is
out-of-domain for likelihood/impact and MUST NOT appear in a logged gap's L×I cells (a gap
so graded is a transcription error, flagged not mapped)" — decided now, stamped with the
mapping version.

### L6-F5 — §3.3(v)'s listing surface claims "every seat" reads it; the lens seats never do

**Severity: LOW — likelihood certain on the textual overclaim (lens dispatch prompt read
first-hand: report.md + CHANGELOG + ledger; debate.md absent) × impact low (actuation-gated;
the cumulative auto-docket is the mechanism-bearing backstop, and the judge's full read is
near-certain in docket-bearing rounds) × complexity trivial. Corroboration: HIGH. Vector
honesty: red's R3-3 fix named "blue's and the judge's read surfaces" — the extension to
"every seat" and the lens's docket right are blue's round-3 additions.**
**Lineage note for the merge: successor-candidate to R3-3.**

**Location:** §3.3 clause (v) — *"pending-entry deltas are LISTED in the round's `### RED`
debate entry — a git-tracked surface every seat and the human operator already read — and
any seat (blue, a lens, the lead, the operator) may docket a listed delta"*.

**Problem:** the read-surface claim is contract-false for lenses: the lens dispatch prompt
names `blue/report.md` (+ CHANGELOG, + citation-ledger); no lens reads debate.md, so a
lens's docket right is decorative. The seats that DO read the round's `### RED` in-run are
blue-respond (the delta's beneficiary — it proposed the deflation) and red-merge (its
author, who just accepted). The window's effective independent watchmen are the judge (in
full, but only when dispatched — near-certain while any docket exists, absent on a clean
board, which is exactly the low-traffic end state actuation is designed for) and the human
operator. The window is materially better than the R3-3 hole — the surface is real and the
cumulative threshold backstops it — but the sentence overstates its watchman set, and an
overstated watchman set is how the R2-6 → R3-3 overclaim survived two rounds.

**Required fix:** restate the reader set honestly ("a surface blue, the judge when
dispatched, and the operator read; lens seats do not") and either add debate.md's latest
`### RED` to the lens read surface in the sharded-era prompt (one line, priced) or drop "a
lens" from the docket-rights list.

### L6-F6 — `realized`-exclusion creates a 3.5-point cliff that rewards under-classification and hides realized-open load from the actuation series

**Severity: LOW — likelihood low-medium (the series is collected in runs 4–5 regardless;
the gradient matters only if the §2.5-item-3 revisit ever runs, which is its designed
purpose) × impact low-medium (biased validation evidence: realized-heavy boards read
artificially low in the exact series the deferred actuation decision keys on; plus a
perverse incentive edge under the carried throttle design — re-grading certain→realized,
evidence STRENGTHENING, cuts board mass by 3.5×weight and with it red's own citation
instances) × complexity trivial. Corroboration: HIGH (report-internal + enum first-hand;
run-3 instance: R4-1 was graded "certain (already realized in this corpus)" — the corpus's
most consequential finding would contribute 0 under the new mapping while sitting open).**
**Lineage note for the merge: successor-candidate to R3-8's Q6 decision text — not a
re-litigation of the exclusion (the lead ruled; the ontological ground is sound), a
completion of its risk grading: the decision was priced on what realized IS, not on what
excluding it DOES to the series and the incentive surface.**

**Location:** §8 Q6 — *"`realized` is EXCLUDED from open-gap mass — realized risk is no
longer a probability; a realized-but-open gap counts in the board profile's open/severity
columns and contributes 0 to mass"* — and §2.5's carried-design note.

**Problem:** two unpriced consequences. (a) **The validation bias:** §2.5 item 3 revisits
actuation "when runs 4–5's logged record shows mass ... actually predicting next-round
value" — but a board carrying realized-but-open gaps now reads low-mass by construction, so
the test's input systematically understates exactly the boards where open work is most
serious; a spurious "low mass predicted low value" pass is cheaper to produce. The
open/severity columns partially mitigate (the profile shows the gap), but the named test
statistic is mass. (b) **The cliff:** certain=3.5 → realized=0 means the boundary between
"certain to bite" and "has bitten" — a word choice in red's own grading prose — moves 3.5 ×
impact-weight of mass. Under the carried throttle design that word choice would set red's
lens budget; no other adjacent enum pair moves more than 0.5.

**Required fix:** one telemetry column and one sentence — add `realized_open` (count, or
excluded-mass memo) to the §2.5-item-1 line so the exclusion never hides realized load from
the record, and note the certain/realized cliff in the mapping pin so any future actuation
review reads boundary re-grades as the sensitive class they are.

### L6-F7 — R3-3's window and R3-7's recompute clause do not compose: a correctly held delta reads as a telemetry discrepancy

**Severity: LOW — likelihood low-medium (requires an actuation review sampling a round with
a pending-window delta — rare but nonrandom: dispute traffic and actuation arrive together
by design) × impact low (a false-alarm discrepancy wastes the review — or worse, the
reviewer "corrects" the telemetry line to match findings, silently defeating the window) ×
complexity trivial. Corroboration: HIGH (both round-3 clauses quoted side by side; the
divergence is by construction). Class: sibling-repair composition — R3-6's exact shape, two
same-round repairs to sibling mechanisms, each faithful in isolation.**
**Lineage note for the merge: successor-candidate to R3-3 and R3-7 jointly.**

**Location:** §2.5 item 1 — *"any actuation review MUST recompute the mass/board columns
for a sample of rounds directly from the git-tracked findings record"* — vs §3.3(v) —
*"an accepted-dispute delta enters the mass computation only after a one-round contest
window."*

**Problem:** when red-merge accepts a dispute, the git-tracked findings record carries the
NEW grade immediately; under the window, the telemetry mass for that round correctly still
uses the OLD grade. A recompute "directly from the findings record" for that round therefore
diverges from a CORRECT telemetry line by exactly the held delta — the re-derivation clause
flags correct window behavior as a transcription discrepancy, and nothing in either clause
tells the reviewer which artifact wins. The failure directions are both bad: treat the
divergence as red-merge error (discredit a sound series) or "fix" the telemetry to match
findings (retroactively actuate the delta the window exists to hold).

**Required fix:** one sentence in §2.5 item 1: the recompute reconciles via the line's own
delta record and the `### RED` listing — a pending-window delta is EXPECTED divergence
(findings shows the new grade, mass holds the old until the window closes), and the
telemetry line with its delta record is correct as logged.

## Checked and deliberately not raised

- Cumulative-threshold residual (≤ threshold−ε per round × rounds after the R3-3(c) fix):
  inherent to any threshold; the report body claims only the 5× kill, which is true; only
  debate.md's round-3 BLUE says "killing the salami arithmetic," and transcript entries are
  not repairable surface.
- awk character-vs-byte semantics in the archive-fraction derivation (105,223 vs 106,772):
  fraction robust — both terms equally affected; noted in the verification block above.
- cost-audit.mjs extension needs a second (runDir) argument its CLI lacks: the report
  already states the current script "cannot consume it and is not cited as if it could";
  spec detail below raise threshold.
- decompose-merge.mjs default RATE=1.0: matches cost.md's cache-read pricing (7.87M tokens
  = $7.87); no discrepancy.
- Terminal-exit judge dispatch (clause (vi)) executor: the script is the executor at build
  time and a post-loop dispatch site is physically constructible; spec coherent as future
  design.
- §3.3(v)'s "checkable against the telemetry line's delta record": both the listing and the
  line share one author (red-merge) — a consistency check within §6.2's stated tier, claimed
  as no more; within ceiling.
- board-telemetry gitignore/glob hazards: re-verified live this round (see verification
  block); premise holds.

## Friction

- 25k Read cap vs blue/report.md again — now 1535 lines / ~56k tokens at round 4, three
  windowed reads for the mandatory full re-read at this lens seat; the audit surface has
  grown ~75% since round 1 and the class recurs at every seat every round (friction #15
  lineage). Same want: a sanctioned whole-audit-surface read mode.

## Synopsis

Seven findings, all lens-scoped: the R3-1 repair's substance VERIFIES at the leaf (instrument
re-run first-hand, every figure reproduces) but its "committed" status word is false in git
terms (L6-F1); the money map compares a corrected $/run figure against an uncorrected $/round
sibling (L6-F2); the R3-11 preflight contradicts skeleton-born and fires after the point of
no return (L6-F3); the pinned mapping omits a schema-legal enum value (L6-F4); §3.3(v)
overclaims its watchman set (L6-F5); realized-exclusion's gradient and validation bias are
unpriced (L6-F6); R3-3 × R3-7 do not compose on window-held deltas (L6-F7). No disposition
flips; top items LOW-MEDIUM.
