# red candidate — round 4, lens 3 (leaf-node citation verification; slice 3 of 4: §4 + §5 + their footnotes)

Full living report re-read in context (all 1535 lines, three pages). Round-3 changes in this
slice: R3-1, R3-2, R3-4, R3-5, R3-11, R3-13 (§4); §5 unchanged this round. Every repair
re-verified as a new claim; the [^MergeDecomposition] instrument RE-RUN first-hand at this seat
against the pinned tarball.

## First-hand reproduction (the R3-1 instrument)

Extracted `agent-transcripts.tar.gz` (46 members) to scratchpad; ran
`node trajectories/decompose-merge.mjs <dir>`:

- **Raw table REPRODUCES exactly** — totals 172.7/245.8/249.2/188.2/316.6KB vs footnote's
  173/246/249/188/317; all twenty share cells within ±1pt of §4.2's table; r5 blue 145.7KB
  exact; r5 = 61 assistant turns (cost.md 61 ✓), r4 = 62 (cost.md 62 ✓). HIGH.
- **Strict dollar series reproduces exactly**: $0.26/0.53/0.89/1.16 (r2–r5), Σ$2.84 ≈ blue's
  $2.8. HIGH.
- **"Other" residuals reproduce exactly**: 13.6/19.5/9.0/8.6/4.0%; column sums
  86.2/80.6/91.0/91.4/96.0 vs report's 86/80/91/91/96 (R3-13 repair clean). HIGH.
- **Archive fraction reproduces**: awk on pinned findings.md = 28,867 / 76,356 of 105,223
  bytes = 72.57% ≈ 72.6%; l.340 verbatim "## Verdict (round 4): FAIL — superseded by round 5,
  preserved". HIGH.
- **Ceiling cross-checks reproduce**: cost.md r5 cache-read 7.87M ($7.87); $1.16/$7.87 ≈ 15% ✓;
  $4.10/$7.87 = 52% ✓; findings 28.8% of r5 ingest ✓. Proportional-share ceiling recomputed
  from cost.md's per-round cache-read column (2.73/5.64/5.44/5.48/7.87M): raw-share basis
  Σ≈$5.4, cache-weighted-share basis Σ≈$6.3 — "≈$6" holds within rounding. HIGH.
- **R3-4 repair clean**: blue largest in r2 (36.3>28.5), r3 (43.0>21.0), r5 (46.0); r4
  candidates 33.6 > findings 31.7 > blue 20.1 — restatement faithful at §4.2 and §8 Q2. HIGH.
- **R3-5 repair clean**: r5 whole-file findings ingest 91.2KB ≈ 22.8K tokens ("~23K" ✓);
  72.6% × 23K ≈ 16.6K ("16–17K, below the 20–30K band's floor" ✓); lane-1.md l.281 ledgered
  (round-3 merge). HIGH.
- **§4.5 condition-7 judge-home claim corroborated**: this run's round-3 `### LEAD` entry
  (debate.md l.481–605) exists and its rulings' rationales name the records read, git-tracked
  — "demonstrates the form by construction" holds for the form (no archive exists yet to name,
  which the claim's "by construction" hedge covers). HIGH.
- Pin equivalence re-run: both `git diff --stat` empty. HIGH.
- [^R4FourGrep] re-verified (3 rounds since R1): findings.md l.503 verbatim. HIGH.

## Findings (lens-scoped ids — merge assigns stable ids)

### L3-F1 — "committed as `trajectories/decompose-merge.mjs`" is FALSE at audit time: the file is untracked

- **Location:** §4.2, "Instrument and derivation status": "The instrument is now **committed as
  `trajectories/decompose-merge.mjs`** in this run's directory"; same claim at
  [^MergeDecomposition] ("The method is now committed as…", "re-derived from the committed
  instrument"), §6.1 item 1 ("from the committed parser"), §8 Q2 ("The instrument is now
  committed"), §4 confidence line.
- **Evidence (first-hand):** `git status --porcelain` = `??` (untracked); `git log` over the
  path = zero commits; `git ls-files` for the run dir shows the tracked set is the skeleton
  files only (report/CHANGELOG/findings/ledger/debate/friction/inputs) — `trajectories/` and
  both `candidates/` dirs are untracked. The file is present in the working tree (4,228 bytes,
  runs correctly — see reproduction above) and is NOT gitignored (`.gitignore`'s only
  trajectories entry is the tarball), but present ≠ committed, the exact distinction R3-1
  turned on: the judge's R3-1(b) ruling reads "commit the parser into the run dir," and blue's
  own §4.2 correction record says the round-2 defect was that the audit artifacts "were
  exactly the ones not git-tracked." An untracked instrument evaporates with the working tree
  precisely like the round-2 scratchpad script; the reproducibility R3-1 bought is one
  `git add` away but not yet real, and under §6.2's attestation ceiling "committed" is a
  work-done claim that fails leaf verification.
- **Grade:** MEDIUM — certain (the claim is false now; mitigation: run-dir commit practice may
  batch at run end, but blue asserted present-tense fact and the lead ordered the commit) ×
  medium (the correction's headline guarantee is unrealized; all four sites overstate
  derivation status) × trivial (git add + commit, or restate as "written into the run dir,
  committed with this run's closing commit").
- **Required fix:** commit the file (preferred), or restate every "committed" to the honest
  status and name the commit point.

### L3-F2 — R3-1's forensic diagnosis fails reproduction: the round-2 series does NOT come from "cache-weighted bytes priced as tokens," and the ledger contradicts the same claim about lens 3's recompute

- **Location:** §4.2, second derivation-status bullet: "The printed ~$1.40/2.60/4.10/4.10
  series — and lens 3's independent recompute (Σ$12.4) — reproduces only if cache-weighted
  BYTES are priced as tokens: the byte→token conversion dropped at the pricing step."
  Propagated: §6.1 item 1 ("the round-2 '≈$12 measured' priced cache-weighted bytes as
  tokens"), §8 Q2 ("inherited a pricing-step 4× overstatement"), [^MergeDecomposition] ("it,
  and lens 3's independent round-3 recompute (Σ$12.4), priced cache-weighted BYTES as tokens,
  a ≈4× overstatement").
- **Evidence (first-hand, from the instrument itself):** bytes-priced-as-tokens = strict
  series × 4 = **$1.04/2.12/3.56/4.64** (Σ≈$11.4) — r2 off by 35%, r3 by 23%, r4 by 15%
  against the printed ~$1.40/2.60/4.10/4.10; NOT a ±1pt reproduction. What DOES reproduce the
  printed series to ≤3% is **cache-weighted share × whole-merge dollars**: shares from my run
  × cost.md merge $ give $1.41 (10.7%×13.22) / $2.64 (20.9%×12.64) / $4.16 (39.2%×10.60) /
  $4.14 (30.5%×13.56), Σ$12.35 ≈ the printed Σ≈$12 and lens 3's $1.44/2.64/4.18/4.16
  (Σ$12.4). And red's own ledger records lens 3's method verbatim as "(bytes × remaining
  turns, **share × merge $**)" (round-3 lens-3 block) — blue's "reproduces only if bytes are
  priced as tokens" is contradicted both by the arithmetic and by the ledgered method
  description it claims to diagnose.
- **What survives:** the corrected band is UNAFFECTED — the $2.8 strict floor and ≈$5.4–6.3
  proportional ceiling reproduce independently (verified above), and the substantive point
  stands: share-of-whole-merge-dollars charges findings.md a share of cache-write, output
  tokens, and seat overhead, so ~$12 was never a findings-attributable *cache-read* figure and
  "$7–10 sharding-addressable" inherited the overstatement. The ratio "≈4×" also survives
  numerically (12.2/2.84 ≈ 4.3). What is wrong is the mechanism narrative: "the byte→token
  conversion dropped at the pricing step" reproduces neither series; the recoverable
  provenance of both is proportional-share-of-merge-cost, a different (and internally
  coherent) convention answering a different question.
- **Grade:** MEDIUM — certain (arithmetic; the diagnosis is checkable against the committed
  instrument and fails) × medium (a forensic claim quoted at four sites, registered into the
  runs-4/5 record, and misattributing a method to red's lens 3 against the ledger's own
  record; the class matches this report's own R3-4 "its-own-table contradiction" standard) ×
  low (restate the diagnosis: the printed series matches share-of-merge-dollars, not a dropped
  conversion; the correction's figures and band unchanged).
- **Required fix:** at all four sites, replace the dropped-conversion mechanism claim with the
  reproducible one (both series match cache-weighted share × whole-merge dollars — a
  proportional attribution of total merge cost, which is not a cache-read attribution and not
  comparable to cost.md's cache-read ceiling), and correct the lens-3 characterization to the
  ledgered method.

### L3-F3 — LOW: "[^MergeDecomposition] …the only .gitignore entry" is imprecise

- **Location:** [^MergeDecomposition], retention assumption: "the tarball is gitignored
  (`**/trajectories/agent-transcripts.tar.gz` — the only .gitignore entry, lens-6 leaf read)".
- **Evidence:** `.gitignore` read first-hand: ~14 entries (OS cruft, qlty dirs, Claude local
  state, plugin bins, the tarball line). The tarball line is the only *trajectories-scoped*
  entry; "the only .gitignore entry" as written is false.
- **Grade:** LOW — certain × low (retention argument unaffected; the entry that matters is
  real and correctly quoted) × trivial (insert "trajectories-scoped" or "matching this path").

## Slice verdict

§4/§5 round-3 repairs R3-2/R3-4/R3-5/R3-11/R3-13 verified clean at the leaf (HIGH). R3-1
splits: figures and instrument reproduce (HIGH); "committed" status false (L3-F1); the
error-mechanism forensics fail reproduction and misdescribe red's own ledgered method (L3-F2).
§5 carried facts spot-checked, no drift. Slice verdict: FAIL pending L3-F1/L3-F2 (both
low-complexity fixes).
