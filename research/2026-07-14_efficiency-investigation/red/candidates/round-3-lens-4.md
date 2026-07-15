# round 3 — lens 4 (leaf-node citation verification), slice 4: §6 Cross-cutting, §7 Pre-flight self-audit, §8 Open questions, Footnotes

Full living report re-read whole (3 windows, all 1365 lines) before auditing; CHANGELOG used
as navigation only. Ledger consulted first; carried-HIGH entries not re-fetched except where
the section changed round 2 or the claim is volatile. All new/amended citations in this slice
leaf-verified this pass; ledger appended (8 entries).

## Verification results — changed/new content in slice

| Claim site | Reference | Result | Confidence |
|---|---|---|---|
| [^Sprt] restored full quotation ("for symmetric error bounds, ... by at least 36% and by at most 75%") | arXiv:2603.00216 abstract, leaf-fetched | EXACT word-for-word match; R2-14 repair clean | HIGH |
| [^AdaptiveStability] amended gloss: "typically 4–7, full range 2–8 across the 22 reported configurations (three JudgeAnything rows 2, 2, 8)", max −1.03%, 10-round baseline, "3 of 22 fall outside, at both ends" | arXiv:2510.12697v1, leaf-fetched twice this pass with full Table-2 row enumeration | ALL VERIFIED. Note for the record: a naive count reads Table 2 as 6 datasets × 4 models = 24; row enumeration totals 22 (the two multimodal benchmarks run only 3 vision-capable models). Exactly 3 rows sit outside 4–7 — JudgeAnything 2, 2, 8 — the remaining 19 all in 4–7. The twice-regressed footnote is now correct in every particular; R2-13 repair clean, verified as a new claim per repair-regression discipline | HIGH |
| [^JournalCheck] claim (a): this run's live journal = lifecycle-only, zero `log()` | wf_5cefd2a4-35f/journal.jsonl, direct read at this seat | Shape verified first-hand: now 50 lines = 28 started + 22 result, grep 'researching\|debate ended' = 0, zero non-lifecycle lines. The footnote's 43/22/21 was a self-labeled mid-run snapshot; counts grew, composition identical. Claims (b)/(c) carried HIGH (round-2 merge, ledger ll.152, 116/131) | HIGH |
| [^MergeDecomposition] "identified by cost-audit.mjs's own 'Red merge, round N' head match" | cost-audit.mjs l.33, direct read | Regex `/Red merge, round (\d+)/` present exactly as described | HIGH |
| [^DalalMallows] via-secondary access note (Höhle 2016, "red citation-ledger line 63, graded HIGH-via-detailed-secondary") | red ledger l.63 cross-check | Matches the ledger record verbatim; R2-18 disclosure discharged | HIGH |
| [^PropagationChains] round-2 amendment (verbatim home blue-researcher.md l.14 ONLY; debate.js l.263 = paraphrase; lens-overrule parenthetical) | round-2 merge first-hand reads, ledger l.156 | Matches in every particular, including the "a round-2 red lens graded that repair HIGH ... overruled by the source read" parenthetical (that lens was this one — accurate) | HIGH (carried + cross-checked) |
| §7 rewritten enumeration: "**four** sources are access-blocked from the verifying seats" | report-wide cross-read + R1 ledger | No fifth attempted-and-blocked source exists in the report; the enumeration is complete over the attempted set. SafeRTS/YooHarman/LostMiddle/FentonOhlsson primaries were never ATTEMPTED — a different category, accepted by red round 1 at knowledge-level MEDIUM-HIGH (ledger l.42); carried acceptance, no new gap. R2-18 repair clean at both §2.2 and §7 | HIGH |
| §7 claim-count echo (~160; ceil(160/40)=4+2=6) and §6.1/§2.4 6-seat rescale (~+$2/round, ~+$10/run) | recompute + this round's own dispatch | $49.48 / 25 lens-rounds ≈ $1.98/lens → ~+$2/round ✓; ~+$10/run at 5 rounds ✓; ceil(160/40)=4 ✓ — and live-corroborated: round 3 dispatched exactly 4 citation instances + 2 fixed lenses | HIGH |
| §6.1 items 3/5, §6.2 attestation ceiling, §6.4 item 6 repricing — engine facts (carried never enters adjudicated ll.252–253; one dispatch/round ll.247–250; judge full-read l.249; lineage throw closures vs gaps[].supersedes; grades machine-readable in redEnv) | debate.js, ledgered HIGH rounds 1–2 (ll.94, 128, 155) | Carried; the round-2 repricing text is consistent with every ledgered mechanism fact | HIGH (carried) |
| §6.3, §6.4 items 1–5, §6.1 cost.md figures, [^PromptCaching] (volatile but same-day) | ledger, rounds 1–2 | Unchanged sections, within the 2-round window, no volatility trigger — carried | as ledgered |

## Findings

### L4-F1 — [^MergeDecomposition]'s measurement is not preserved for independent re-derivation — NEW

- **Location:** Footnotes, `[^MergeDecomposition]`: "each red-merge transcript ... parsed with
  a ~70-line node script" and "method stated here so run 4 can reproduce it in one parse";
  load-bearing consumers: §4.2's decomposition table, §6.1 item 1 ("measured round 2: ≈$12/run
  of merge cost is findings-attributable, ~$7–10 sharding-addressable"), §8 Q2 ("ANSWERED for
  run 3").
- **The gap:** the parser script exists nowhere in the run directory (glob `**/*.{mjs,js}`
  under the run dir = zero hits) — it lived and died in the session scratchpad — and its input
  (`agent-transcripts.tar.gz`) is untracked working-tree material that a cleanup can delete.
  The §4.2 table is therefore blue self-report whose re-derivation path is "rewrite the parser
  from the footnote's prose while the untracked tarball still survives." No red seat has
  re-derived the shares (round-2 merge verified the tarball's existence, size, and member
  count — not the decomposition numbers). Under the report's own §6.2 attestation ceiling,
  vacuity/error in a work-done claim is checked post-hoc by independent seats **over
  git-tracked artifacts** — this measurement's artifacts are exactly the ones that aren't.
- **Grading:** likelihood low-medium (the numbers are plausible against known file sizes and
  blue's method prose is competent; but transposition/weighting errors in a one-shot
  never-re-run script are the base-rate case this run's own §6.4 exists to document) × impact
  medium (the money map's #1 ranking and the ~$7–10 sharding-addressable figure feed the run's
  headline disposition; §4 confidence already honestly capped at MEDIUM-HIGH) × complexity
  trivial (commit the ~70-line script into the run dir — e.g.
  `trajectories/decompose-merge.mjs` or inline in an appendix — and state the tarball-retention
  assumption). **LOW-MEDIUM overall.**
- **Required fix:** preserve the parser as a tracked artifact of this run and name the
  tarball's retention status in the footnote; alternatively re-derive one round's shares at an
  independent seat and ledger it. Not a corroboration failure — a preservation gap on a
  measurement the report leans on.

## Lens verdict (slice 4)

Slice is clean apart from L4-F1 (LOW-MEDIUM, preservation not corroboration). All three
round-2 repairs landing in this slice ([^Sprt] R2-14, [^AdaptiveStability] R2-13, §7/
[^DalalMallows] R2-18) leaf-verify exactly — the AdaptiveStability footnote, twice regressed
in earlier rounds, is now correct in every checkable particular including the "22
configurations" count, which survives a row-level enumeration a naive table read would get
wrong. No repair-regression instances found this round in this slice.

friction: none — all sources fetchable; the second AdaptiveStability fetch (row enumeration)
was needed because the first fetch's summarizer asserted 6×4=24 by arithmetic rather than
counting rows; no tool gap, just a re-prompt.
