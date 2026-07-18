# Red round-2 — Lens 3 (leaf-node citation verification)

**Slice:** §5 (H5 cost discipline) + §6 (risk matrix) + owned footnotes
(`[^CliReference]`, `[^HeadlessProbe]`, `[^HeadlessDocs]`, `[^UsageAPI]`,
`[^RateLimitsAPI]`, `[^ConsoleLimits]`, `[^Pricing]`, `[^CostRecord]`,
`[^EfficiencyPlan]`, `[^EffReport]`, `[^ResearchCommand]`, `[^FrictionRun4]`,
`[^RoutinesDocs]`, `[^McpHeadlessBugs]`, `[^HooksJson]`, `[^PushGuard]`,
`[^QmdFallback]`).

**Method note.** No Round-2 blue CHANGELOG entry exists → §5/§6 content is byte-stable
since the round-1 revision. Every §5/§6 external leaf was verified HIGH last round at
access date 2026-07-17 (= today); zero time has elapsed, so living-source drift is
impossible on the clock. Per the ledger discipline I did NOT re-fetch the already-HIGH
docs/issue leaves. I DID re-fetch the two leaves in my slice that carried an unresolved
grade or an explicit re-fetch directive: the pricing page (`[^Pricing]`, flagged
"VOLATILE — re-fetch at citation-verification") and the never-fetched batch-processing
page (the standing MEDIUM ≤24h sub-claim). Both are citation-verification duties this lens
owns.

---

## Positive verifications (no gap; ledger upgrades)

**V1 — Pricing figures re-confirmed HIGH, no drift.** Live re-fetch of
`platform.claude.com/docs/en/about-claude/pricing` (2026-07-17) corroborates §5.2 verbatim:
Haiku 4.5 $1/$5; Sonnet 4.5/4.6 $3/$15; Sonnet 5 intro $2/$10 through 2026-08-31 → $3/$15
from 2026-09-01; Opus 4.5–4.8 $5/$25; Fable 5 & Mythos 5 $10/$50; Batch flat 50% off; cache
read 0.1× (= ~90% cut on cached input). The +30% tokenizer note is verbatim on the page:
"Claude Opus 4.7 and later Opus models, Claude Fable 5, Claude Mythos 5, Claude Mythos
Preview, and Claude Sonnet 5 use a newer tokenizer ... produces approximately 30% more
tokens for the same text ... Claude Sonnet 4.6 and earlier models use the previous
tokenizer." §5.2's tokenizer caveat is accurate, including the boundary (Sonnet 4.6 =
legacy).

**V2 — Batch ≤24h window now CORROBORATED (MEDIUM → HIGH available).** Blue carried the
≤24h Batch async-window sub-claim MEDIUM in `[^Pricing]`, naming the batch-processing page
as the pending check. Live fetch of `.../build-with-claude/batch-processing` (2026-07-17)
states it directly: "You can access batch results when all messages have completed or after
24 hours, whichever comes first. **Batches expire if processing does not complete within 24
hours.**" and "plan your batch submissions with the 24-hour processing window in mind."
This is not a gap against blue (honestly labeled + exact page named); it is a bankable
upgrade — the sub-claim can move to HIGH, and the standing "unable to corroborate" is
retired. Attempt-line satisfied: fetched, quote captured. (Stakes are low regardless —
Batch is a demoted FUTURE note per R1-23.)

---

## Findings

### L3-F1 (LOW) — R1-9 requalification did not propagate to §6 risk-matrix row 5

**Location:** §6 risk matrix, row 5 — "No programmatic quota pre-check | **Certain (no
API)** | Low ... | RISK-ACCEPT with §5.1's layered static guards".

§5.1 was carefully requalified in round 1 (R1-9): spend limits have no API read/set, but
**rate limits ARE API-readable** (`[^RateLimitsAPI]`, `/v1/organizations/rate_limits`),
merely unreachable by a subscription-auth scheduler. Blue's own CHANGELOG lists R1-9 edits
touching "§5.1, H5 verdict, [^ConsoleLimits], new [^RateLimitsAPI]" — **not** §6 row 5. Row
5's likelihood cell still reads the pre-R1-9 flat "(no API)", which §5.1 now explicitly
contradicts as over-broad ("RATE limits, by contrast, ARE API-readable").

**Grading:** likelihood × impact × complexity all unaffected — the *conclusion* (a
subscription-auth scheduler cannot poll any quota pre-check) is unchanged and correct; only
the parenthetical justification inside the cell is stale. This is a propagation-completeness
miss of the incomplete-repair class, not a substantive error. **LOW** (likelihood-of-harm
negligible; it is a cosmetic staleness a reader cross-checking §5.1 will notice). Suggested
fix: change row 5's "(no API)" to "(no spend-limit API; rate-limit API unreachable at this
auth tier — §5.1/R1-9)". Corroboration confidence on the underlying §5.1 claim: HIGH
(`[^RateLimitsAPI]`/`[^UsageAPI]` verified round 1, unchanged).

### L3-F2 (LOW) — §5.2 run-3 figure carries a bundled footnote whose sub-source differs from the body marker

**Location:** §5.2 — "a full debate run cost **$414.97 at list rates** (...); run 3 was
$149.95 [minority: lane-3/local-probe — the run-3 figure].[^CostRecord]".

The body attaches a single `[^CostRecord]` marker to two figures with different sources:
$414.97 is from `cost.md`, but $149.95 is from `plans/efficiency-phase.md §I` — as the
`[^CostRecord]` footnote itself discloses ("Run-3 baseline $149.95 per
plans/efficiency-phase.md §I"). Footnote-over-attribution pattern (one marker, a bundled
claim-list, only part of which the named artifact carries). **Net LOW** because the footnote
is honest about the split and the $149.95 figure is independently verified HIGH
(`[^EfficiencyPlan]` line 6, round-1 ledger). No reader is misled who follows to the
footnote. No fix strictly required; if touched, add `[^EfficiencyPlan]` beside the run-3
figure. Corroboration confidence: HIGH on both figures.

---

## Slice verdict

§5/§6 citation surface is **clean at HIGH**. No new mint-worthy gap; the two findings above
are LOW (propagation-staleness and a disclosed bundled footnote), neither altering any
grade or verdict. One bankable ledger upgrade (V2, Batch ≤24h → HIGH). All external leaves
in the slice re-confirmed or held at their prior HIGH grade with no drift (same access
date). MUST-TRY observable satisfied: pricing + batch-processing pages fetched live, quotes
captured; no leaf left at "unable to corroborate".
