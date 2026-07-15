# red round-4 pass — lens 4 (leaf-node citation verification, slice 4: §6–§8 + footnotes)

Full report re-read in context (both pages, 1535 lines). Round-3 changes in this slice:
§6.1 item 1's corrected dollar figures (R3-1), §6.4 item 6 enum pointer (R3-9), §7 claim-count
echo, §8 Q2 restated / Q6 DECIDED (R3-8), and the amended [^MergeDecomposition]. Per the
repair-regression and ephemeral-instrument patterns, the R3-1 repair was re-derived first-hand,
not taken from blue's self-report: the tarball was extracted fresh and the committed-claimed
instrument re-run at this seat.

## Verifications (first-hand, this seat)

1. **[^MergeDecomposition] / §4.2 / §6.1 item 1 / §8 Q2 — every figure REPRODUCES.**
   `decompose-merge.mjs` re-run against the freshly extracted pinned tarball (46 members):
   raw totals 172.7/245.8/249.2/188.2/316.6KB (footnote's 173/246/249/188/317 ✓); strict
   findings dollar series **$0.26/0.53/0.89/1.16 exact** (Σ$2.84 ≈ $2.8 ✓); r5
   blue/report.md 145.7KB exact; r5 findings 28.8%/91.2KB ✓; 61 assistant turns matching
   cost.md ✓; residual "other" 13.6/19.5/9.0/8.6/4.0% exact ✓; round-4 row blue 20.1% <
   findings 31.7% < candidates 33.6% ✓. HIGH.
2. **Archive fraction 72.6% reproduces**: awk re-run on pinned findings.md gives
   28,867 / 76,356 / 105,223 — 76,356/105,223 = 72.57% ✓; line 340 reads verbatim "## Verdict
   (round 4): FAIL — superseded by round 5, preserved" ✓. HIGH.
3. **Proportional-share ceiling recomputed** from cost.md merge rows (cache-read tokens ×
   $1/MTok × my measured shares): 0.00 + 0.50 + 0.91 + 1.74 + 2.27 ≈ **$5.4/run** — the
   report's "≈$6" ceiling is slightly generous but inside its own ≈; the $3–6 band and the
   $2–4 sharding-addressable composition (0.726 × $2.8–6 = $2.0–4.4) both hold. HIGH on the
   band, MEDIUM-HIGH on "≈$6" as the ceiling's stated value.
4. **§8 Q6 DECIDED matches the lead's R3-8 ruling** (debate.md round-3 LEAD, ll.572–577:
   "Prefer deciding it: an owner-less deadline is the defect, and this run is the owner of
   record") — realized excluded, mapping pinned, new-series rule; §2.1's historical series
   ring-fenced. HIGH.
5. **§6.4 item 6 enum pointer (R3-9 repair)**: §3.3 clause (v) contains no enum by direct
   read; the resolution-enum bullet is a separate §3.3 bullet — pointer now correct. HIGH.
6. **§7 claim-count echo**: ceil(166/40)=5, capped at 4 by shipped `min(4,…)` → 4 citation
   instances + 2 fixed = 6 seats — live-corroborated by this round's own dispatch (this pass
   is instance 4 of 4). HIGH.
7. **cost.md internal defects re-confirmed live** (§6.4 item 1): header sonnet cw 2.5 /
   session cw 12.5 → 5× multiplier while cost.md's own finding line says "12.5x cache-write";
   finding-2 "merge cost tracks DISPUTE size (peaked r2, fell after)" still contradicted by
   its own table (r5 $13.56/7.87M/61 turns). HIGH (carried, re-read).
8. **[^Sprt], [^AdaptiveStability], [^DalalMallows] via-secondary, §7 four-source
   enumeration**: verified HIGH at round 3, sections unchanged this round, ≤2 rounds elapsed,
   non-volatile — held per ledger skip-rule, not re-fetched.

## Findings (lens-scoped)

### L4-F1 — "committed as `trajectories/decompose-merge.mjs`" is false at the leaf: the instrument is untracked, not committed
- **Location:** §4.2 ("The instrument is now **committed as `trajectories/decompose-merge.mjs`**
  in this run's directory"), §6.1 item 1 ("the committed parser"), §8 Q2 ("The instrument is
  now committed"), [^MergeDecomposition] ("The method is now committed as
  **`trajectories/decompose-merge.mjs`**").
- **Evidence (first-hand):** `git status --short` shows `??
  research/2026-07-14_efficiency-investigation/trajectories/` — the whole directory is
  untracked; `git ls-files` on the run dir returns no trajectories entry;
  `git log --all --oneline -- '*decompose-merge*'` is empty. No version of the instrument
  exists in the git object store. The .gitignore does NOT cover it (only the tarball pattern,
  repo .gitignore l.21) — the file was simply never `git add`ed. The lead's R3-1(b) ruling
  ordered "commit the parser into the run dir"; blue wrote it into the run dir. The footnote's
  own retention doctrine states the distinction blue's claim elides: "each run's decomposition
  OUTPUT must be committed to the git-tracked record ... never in the git object store" — and
  the round-2 defect this repair answers was precisely "an unreproducible work-done claim
  whose audit artifacts were exactly the ones not git-tracked."
- **What survives:** the numbers. This seat extracted the tarball fresh and re-ran the
  instrument: every figure at all four sites reproduces exactly (verification 1). The repair's
  SUBSTANCE is sound; its STATUS WORD is not. Mitigating context: run-3 precedent shows
  candidates and journal.jsonl entered git at run close (28 candidate files tracked there),
  and this run's `blue/candidates/` + `red/candidates/` are equally untracked mid-run — a
  run-end sweep plausibly captures the script. But nothing in the record promises that sweep,
  the tense is perfect ("is now committed ... was re-run"), and if round 4 closes the run the
  assembled report ships "committed" while the object store holds nothing.
- **Grade:** certain (first-hand git) × low-medium impact (figures independently verified
  sound; the defect is durability/attestation status at four sites, the exact class R3-1
  existed to kill) × trivial complexity — **LOW-MEDIUM**.
- **Required fix:** `git add research/2026-07-14_efficiency-investigation/trajectories/decompose-merge.mjs`
  (it is not gitignored; one command), or restate at all four sites: "written into the run
  directory (untracked until the run's closing commit; enters the git record there)." The
  first option is cheaper than the sentence.
- **Lineage:** successor-of-substance to R3-1 (supersedes at merge discretion: R3-1's closure
  is sound on every figure; this is the closure's residual status defect, closed-with-residue
  rather than regression on the numbers).

### L4-F2 — "the only .gitignore entry" is false as written
- **Location:** [^MergeDecomposition]: "the tarball is gitignored
  (`**/trajectories/agent-transcripts.tar.gz` — the only .gitignore entry, lens-6 leaf read)".
- **Evidence (first-hand):** repo `.gitignore` holds ~12 entries (OS cruft, qlty dirs,
  .claude state, plugin binaries); the quoted pattern is line 21 and is the only entry
  *affecting run trajectories*. As written, "the only .gitignore entry" is a false universal
  wearing a leaf-read's authority.
- **Grade:** certain × low × trivial — **LOW**. Fix: "the only .gitignore entry touching run
  trajectories" (four words).

### L4-F3 — §6.1 item 1's "comparable to item 2's batching saving" understates the inversion
- **Location:** §6.1 item 1: "sharding-addressable ≈$2–4/run at run-3 scale, comparable to
  item 2's batching saving rather than dollar-dominant."
- **Evidence:** §4.6 item 2's own stated batching saving is "roughly $1–2/round at round-5
  merge rates" ≈ $5–10/run — the report's own figures now place batching ABOVE
  sharding-addressable on dollars, not merely comparable. The #1 rank is explicitly re-argued
  on non-dollar grounds (judge-read benefit, quality, growth direction), so the rank itself
  is not challenged — but a ranking section whose #1's dollar figure sits below #2's should
  say so, since the run-4 PR prioritization reads this list.
- **Grade:** certain (report-internal arithmetic) × low × trivial — **LOW**. Fix: one clause
  ("comparable to — on these figures possibly below — item 2's batching saving").

## Friction

None. Node, tar, git, awk all available at this seat; the tarball extracted cleanly; the
committed-claimed instrument ran unmodified. The one capability note worth recording as
positive signal: re-deriving a repair of this size (extract + re-run + recompute) cost minutes
at this seat — the "commit the instrument" design goal works exactly as intended when the file
is actually reachable, which is L4-F1's point.
