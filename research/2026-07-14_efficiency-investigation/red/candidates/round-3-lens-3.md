# red round-3 — lens 3 (leaf-node citation verification), slice 3: §4 + §5 + carried footnotes

Full living report re-read whole (lines 1–1365) before auditing. Slice: §4 (sharded findings +
collator), §5 (round-scoped audit), and their carried footnotes ([^MergeDecomposition],
[^JournalCheck]-adjacent uses in §4, [^FrictionFour/Ten/Fifteen/Six], [^R4FourGrep],
[^PropagationChains], [^CitationLedgerRun3], [^SafeRTS]/[^YooHarman]/[^DiffReview]/
[^FentonOhlsson] as used in §5.3). Round-2 edits in slice: R2-3 (§4.2 measurement + table +
[^MergeDecomposition]), R2-4 (§4.5 cond 7), R2-5 (§4.5 cond 2), R2-16 (§4.6 item 1), R2-17
(§4.5 cond 6), R2-2 sweeps (§4.6 item 2, §5.5). §5.1–§5.4 unchanged this round; their claims
are ledgered HIGH within the 2-round window and carry.

## Independent re-derivation of the §4.2 decomposition (first-hand)

Extracted `agent-transcripts.tar.gz` (46 members re-confirmed) to the session scratchpad;
identified the five red-merge transcripts by the same `"Red merge, round (\d+)"` head match
cost-audit.mjs uses (l.33, verified); recomputed tool-result bytes per source class AND the
cache-weighted (bytes × remaining assistant turns) dollar attribution:

| round | total (blue's) | blue-report | findings | candidates | debate |
|---|---|---|---|---|---|
| 1 | 174KB (174) | 35.7% (36) | 0.1% (0.1) | 46.6% (46) | 4.0% (4) |
| 2 | 248KB (247) | 36.2% (36) | 8.9% (9) | 28.5% (28) | 6.9% (7) |
| 3 | 252KB (250) | 43.0% (43) | 16.7% (17) | 21.7% (21) | 10.4% (10) |
| 4 | 190KB (190) | 20.0% (20) | 31.8% (32) | 33.6% (33) | 6.0% (6) |
| 5 | 320KB (318) | 45.9% (46) | 28.9% (29) | 19.0% (19) | 2.3% (2) |

Cache-weighted findings dollars: $0.01 / $1.44 / $2.64 / $4.18 / $4.16 (blue: ~$1.40 / ~$2.60 /
~$4.10 / ~$4.10, Σ≈$12) — reproduced within rounding, Σ$12.4. Round-5 blue/report ingest
146.9KB (blue: "145KB"). Lens-candidate ingest 55–81KB by my compute (blue "52–80KB" — matches
its own table's rounded cells). debate.md 2–10% ✓. Findings 60.5/92.4KB rounds 4–5 ≈ 15–23K
tokens at 4 B/tok ✓. **The table, the dollar attribution, and the method's head-match claim are
VERIFIED HIGH first-hand.** The two prose glosses wrapped around the verified table are not —
findings L3-F1, L3-F2 below. The within-file archive split is the one sub-figure the stated
method cannot produce — L3-F3.

## Findings (lens-scoped ids; merge assigns stable ids)

### L3-F1 — §4.2's "largest every round from round 2" gloss is contradicted by its own table at round 4 (and propagated to §8 Q2)

- **Location:** §4.2 "Convergent: the cost case, measured," sentence: "Two bonus measured
  facts: **blue/report.md is the largest merge-context component every round from round 2**
  (145KB ingested at round 5 — run-3 friction #15's real referent, untouched by lever 4a)".
  Propagated site: §8 Q2 — "blue/report.md is the largest merge component every round from
  round 2".
- **Evidence:** the report's OWN table, same section, round-4 row: blue/report.md 20% vs lens
  candidates 33% and red/findings.md 32%. Independently recomputed first-hand from the
  extracted transcripts: raw ingest r4 = blue 20.0% / candidates 33.6% / findings 31.8%;
  cache-weighted r4 = blue 14.9% / findings 39.5% / candidates 38.6%. Under BOTH measures,
  round 4's largest component is not blue/report.md — it is third. True rounds 2, 3, 5 only.
- **Corroboration confidence:** HIGH (contradiction; static text vs recomputed record).
- **Grade:** MEDIUM — certain (recomputed) × medium-low impact (a "bonus measured fact" and
  §8's carried open-question text misstate the measured record at two sites; no disposition
  flips — §4's RATIFY rests on the findings share, which is real) × trivial fix.
- **Required fix:** restate faithfully at both sites, e.g. "the largest merge-context
  component in three of the four rounds from round 2 (all but round 4, where findings.md and
  the lens candidates each exceeded it)." Class: over-universal gloss on a correct table
  (exhaustive-sweep-omits-hard-case / sibling-halo family — the verified table lends the
  false universal its halo).

### L3-F2 — "landing INSIDE lane 1's 20–30K directional estimate" conflates whole-file ingest with lane 1's archive-share estimate, and misses the band at round 4

- **Location:** §4.2, sentence: "(~60–92KB/round by rounds 4–5 ≈ 15–23K tokens at ~4
  bytes/token — landing INSIDE lane 1's 20–30K directional estimate, now measured)".
- **Evidence:** lane-1.md l.281 (leaf-read): "the **archive's share** of merge context is
  maybe 20–30K tokens **by round 5**." Lane 1 estimated the archive sub-fraction at round 5;
  §4.2's measured 15–23K is the WHOLE findings file across rounds 4–5. Round 4's 15K is below
  the band even whole-file; at round 5 the whole file is 23K, so the archive sub-fraction
  ("clear majority" per the same paragraph ⇒ ~14–20K) sits at or below the band's floor. The
  correct comparison shows lane 1 likely OVER-estimated — the opposite of "landing inside."
  (Contrast: the same paragraph's "$7–10 ... squarely in lane 1's $5–15 band" checks out —
  lane-1.md l.283, $7–10 ⊂ $5–15, verified.)
- **Corroboration confidence:** HIGH on the mismatch (both texts read first-hand).
- **Grade:** MEDIUM — certain × low-medium impact (validation rhetoric for the measurement;
  headline $7–10 unaffected) × trivial fix. Class: metric conflation across a band's
  endpoints (ledgered pattern — two different quantities wearing one comparison).
- **Required fix:** either compare like with like ("lane 1's archive-share estimate is
  bounded above by the measured 23K whole-file round-5 ingest — the estimate's upper half is
  excluded, its lower edge plausible") or drop the "landing inside" clause.

### L3-F3 — the sharding-addressable $7–10 rests on a within-file archive/open split the documented method cannot produce

- **Location:** §4.2, sentence: "of which the archive fraction (closed blocks, the
  sharding-addressable part) is the clear majority by late rounds: sharding-addressable ≈
  **$7–10/run at run-3 scale**"; and [^MergeDecomposition]'s reproduction promise: "method
  stated here so run 4 can reproduce it in one parse."
- **Evidence:** [^MergeDecomposition]'s stated method attributes tool-result bytes "to source
  files via their tool_use `file_path`/command" — a per-FILE attribution. No documented step
  measures closed-block bytes WITHIN findings.md, yet the ranked-#1 money-map figure
  (§6.1 item 1 cites "~$7–10 sharding-addressable") and §4's confidence upgrade both ride
  that split. I reproduced every other figure in §4.2 first-hand from the stated method; this
  one sub-figure is not derivable from it. The §4.2 caveat list ("single run; bytes→tokens ≈
  4:1; the weighting ignores the system prompt") does not name the undocumented split.
- **Corroboration confidence:** figures $12/run findings-attributable HIGH (reproduced);
  archive-majority fraction and hence $7–10 LOW-MEDIUM (assertion; direction plausible from
  run-3 file structure — closed blocks dominate late-round findings.md bytes — but the
  claimed status is "measured," and it is not measured by the stated method).
- **Grade:** MEDIUM — high likelihood (the method text demonstrably omits the step) × medium
  impact (it is the sizing of the report's top-ranked lever and the §4 MEDIUM-HIGH confidence
  rationale; run-4's promised re-measurement cannot reproduce the headline from the footnote)
  × low fix complexity.
- **Required fix:** document how the archive fraction was obtained (e.g. closed-block byte
  share of findings.md at each round boundary from the pinned file, or Read-offset windows)
  and mark the $7–10 with its actual derivation status ("measured $12 findings-attributable ×
  estimated archive fraction ~60–80%"), or re-derive it observably.

## Repair verifications (round-2 fixes in slice, leaf-checked)

- **R2-3 (§4.2):** the measurement is real, faithful, and independently reproduced (table,
  dollars, 145KB, 46 members, head-match method) — repair VERIFIED except the two glosses
  above (L3-F1, L3-F2), which are round-2-minted regressions inside an otherwise sound
  repair (repair-regression class: re-verify every repair as a new claim — it held for the
  table, failed for the glosses).
- **R2-17 (§4.5 cond 6):** harness.mjs quotes verbatim first-hand — l.7 "exercised with
  canned envelopes and no live agents", l.24 "A stub world." VERIFIED HIGH.
- **R2-4 (§4.5 cond 7):** lineage-throw contrast (closures vs gaps[].supersedes,
  debate.js ll.227–235) and hook fs-capability both ledgered HIGH ≤2 rounds; rewrite is
  consistent with the pinned code. VERIFIED (carried + spot-checked).
- **R2-5 (§4.5 cond 2):** judge full-read prompt l.249, contested filter ll.244–245,
  findings.md hardcoded at l.216 AND l.249 — ledgered HIGH (rounds 1–2, first-hand at merge
  seats). Carried.
- **R2-16 (§4.6 item 1):** hooks.json `Write|Edit` matcher + later-sweep hazard — matches my
  own round-2 L3-F2 finding and its ledgered repair. VERIFIED (carried).
- **R2-2 sweeps (§4.6 item 2, §5.5):** "6 lens seats/round" from ceil(160/40)+2 — ledgered
  formula + §7's claim-count echo; "52–80KB/round" reproduced. VERIFIED.

## Slice verdict

§4/§5 citations at the leaf: everything reproducible was reproduced and holds; three defects,
all round-2-minted or round-2-surfaced, all trivial-to-low fixes, none disposition-flipping.
No unresolved holds; no conflicts deferred to merge except none. Findings above go to the
merge for stable ids.

## Friction

- Bash tool mangled a quoted heredoc: `<<'EOF'` still ate one backslash of `/\\/g` in a node
  script (became `/\/g`, syntax error) — quoted heredocs are supposed to be literal. Detoured
  via the Write tool. The write-path detour itself then tripped the read-before-write guard
  on a file the broken heredoc had created. Cost: two extra turns. Wanted: heredoc
  passthrough that is actually literal, or a documented "always use Write for scripts" rule.
