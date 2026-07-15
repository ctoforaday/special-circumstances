# Round 3 — lens 1 (leaf-node citation verification), instance 1 of 4

Slice 1: preamble + §0 + §1 + §2 (+ their footnotes: [^PinCheck] [^BacklogLevers] [^FindingsBoard]
[^CostAudit] [^StopResume] [^R4OneDetail] [^R5FiveDetail] [^R5FiveSingleton] [^Round3Red]
[^CeilingDisposition] [^AdaptiveStability] [^DebateRounds] [^DalalMallows] [^Stads]
[^CaptureRecaptureEval] [^CaptureRecaptureDecade] [^Sprt] [^Iso29119] [^CvssInconsistent]
[^ConflictingScores] [^CathedralBazaar] [^ExpertCvss] [^RbtTaxonomy] [^JournalCheck]
[^FrictionTen] [^EngineSource] [^AlreadyShipped]). Full report re-read whole (1365 lines, three
pages). Changed-this-round sites in slice, per CHANGELOG round 2: §0 rows 1–2, §1.2 (R2-10),
§1.5(1) (R2-1), §2.2 (R2-12/R2-18), §2.3 (R2-14), §2.4 (R2-2), §2.5 items 1–2 (R2-1/R2-7),
[^AdaptiveStability] (R2-13), [^Sprt] (R2-14). Every repair re-verified as a new claim
(repair-regression discipline).

## Re-verifications performed this round (leaf-node)

1. **Pin equivalence** — re-run first-hand: `git diff --stat bfa8a3b HEAD -- research/2026-07-12_feov-retrospective/`
   empty; `git diff --stat 5396952 HEAD -- ideas/backlog.md plugins/frank-exchange-of-views/`
   empty. Preamble claim holds. **HIGH.**
2. **[^AdaptiveStability] as amended round 2 (R2-13)** — re-fetched arXiv /html/2510.12697v1.
   Table 2 hand-count of the enumerated rows = 4 (BIG-Bench) + 4 (JudgeBench) + 4 (LLMBar) +
   4 (TruthfulQA) + 3 (MLLM-Judge) + 3 (JudgeAnything) = **22 configurations** — footnote's "22
   reported configurations" verified; exactly **3 rows outside 4–7** (JudgeAnything 2, 2, 8 —
   both ends, as the footnote says); "typically 4–7" fair (19/22); max delta **−1.03%**
   (Gemma-3-4B / BIG-Bench) confirmed; criterion **Dt < 0.05 for 2 consecutive rounds** (KS)
   confirmed. Method note: the fetch model's own summary said "18 configurations" while its
   row list enumerated 22 — the enumeration is the evidence; recounted by hand. **HIGH.**
3. **[^Sprt] restored quotation (R2-14)** — re-fetched arXiv abs/2603.00216. Abstract contains,
   character-faithfully: "for symmetric error bounds, the sequential test reduces the average
   sample size by at least 36\% and by at most 75\%" — condition prefix in the same sentence,
   second "by" present. §2.3's "type-I = type-II" gloss of "symmetric error bounds" is
   faithful. **HIGH.**
4. **[^JournalCheck] clause (a) — the live-journal claim (R2-1)** — verified first-hand at
   this seat: `.../subagents/workflows/wf_5cefd2a4-35f/journal.jsonl` = 50 lines = 28
   `"type":"started"` + 22 `"type":"result"`, **zero other event types, `researching` grep = 0**.
   Blue's mid-run read (43 = 22+21) is consistent with subsequent growth; the composition
   claim — lifecycle events only, no `log()` output — holds exactly. Clauses (b)/(c) (run-3
   journal 87 = 46+41; cost-audit.mjs glob `agent-*.jsonl` l.28, zero journal refs) already
   verified first-hand at the round-2 merge (ledger ll.152). **HIGH.**
5. **§1.2 arithmetic (R2-10)** — recomputed: ~$78 gross (round 3 $25.08 = 2.98+9.46+12.64,
   ledgered R2 + rounds 4–5 Σ$53.00) − ~$10 judge round = **~$68 net** ✓; "all ~21 rounds-3–5
   mints (10+5+6)" = 21 ✓; series-total consistency: 20+11+10+5+6 = 52 = 31 (closed r1–2,
   cost.md) + ~15 (closed r3–5) + 6 (open at ceiling) ✓. **HIGH.**
6. **§2.4 rescale (R2-2)** — recomputed: red-lens $9.22–$11.05/round at 5 agents → ~$2/lens;
   4→1 citation instances = 3 agents ≈ $6/round × 3 late rounds = **~$18/run** ✓; "~10% of the
   rescaled run-4 baseline" consistent with ~$160 rescaled + ≥1–2 judge dispatches (~$180);
   "measured lens-candidate ingest 52–80KB/round" recomputed from §4.2's table (46%×174KB=80,
   28%×247=69, 21%×250=52.5, 33%×190=62.7, 19%×318=60.4) ✓. **HIGH** on arithmetic; savings
   estimates stay MEDIUM as self-graded.
7. **§2.2 repairs (R2-12, R2-18)** — direct read: throttle-input subject restored ("But **the
   throttle input** was measured-robust..."); "genuinely paywalled" replaced by "inaccessible
   from the verifying seats (403 at the abstract; ... bot-block vs paywall unshown)". Both
   present as claimed. CathedralBazaar figures carried (three independent round-1/2 fetches,
   ledger ll.102/113/124). **HIGH.**
8. **§0 rows 1–2 (R2-1 propagation)** — durable-sink conditionality present in both rows;
   consistent with §1.5(1) and §2.5 item 1. Row 3a's conditional-vote discount intact
   (conditional-vote-laundering class does not recur). §0 rows 4a ("seven named conditions" —
   §4.5 has exactly 7), 4b, 5 consistent with their sections. **HIGH.**
9. **§2.1/§2.3 spot recomputes** — mapping products: certain×low = 3.5 vs medium×medium = 4 ✓;
   round-5 mass 31.0 carried (ledger l.77). Carried-claim set for unchanged §1.1/§1.3/§1.4
   (board table, backlog quotes, Round3Red, StopResume, CeilingDisposition, DalalMallows-via-
   secondary, Stads, capture-recapture pair) all ledgered HIGH/MEDIUM-HIGH round 1, sections
   unchanged, within 2 rounds — carried per ledger discipline.

## Findings (lens-scoped ids)

**L1-F1 — LOW (likelihood certain-as-written × impact low × complexity trivial).**
Location: §2.1 "Convergent: the computed mass series, and its correlation failure", item 4 —
"mean mass per gap stays ~5 across all five rounds because late-round textual nits carry
`certain` likelihood by construction". Recomputed from the report's own lane-1 table:
98/20=4.9, 65/11=5.9, 44/10=4.4, 30/5=6.0, 31/6=5.2 (lane 3: 4.9, 5.6, 4.4, 5.8, 5.3). The
band is 4.4–6.0 — "stays ~5" stretches ~20% at the round-4 end. The argumentative point
survives and is arguably *strengthened* (late-round means 6.0/5.2 are the highest/third-highest
— the metric shows no downward trend on trivia-heavy rounds), so this is a precision nit, not
a load-bearing defect. Required fix: state the band ("mean mass per gap sits in a flat
4.4–6.0 band with no downward trend") or keep "~5" with the range parenthesized. Round-0 text,
unchanged section, >2 audit passes old — flagged now under the ledger's elapsed-rounds rule.

No other new gaps in slice 1. All eight round-2 repairs landing in this slice verified at the
leaf with zero regressions — first round in this run a repair batch has come back clean in
this slice.

## Verdict contribution (slice-scoped)

PASS-leaning for slice 1: one LOW precision nit (L1-F1); no MEDIUM+ candidates. Merge holds
the run verdict.

## Friction

WebFetch extraction model produced an internally inconsistent read of arXiv 2510.12697 Table 2
— its prose summary asserted "18 reported configurations" while its own row enumeration listed
22; the correct count (22) required hand-recounting the enumerated rows. The lossy-PDF/HTML
extraction class again: summaries from the fetch model are not evidence; only its quoted
enumerations are. A table-faithful extraction path (arxiv-latex MCP) was not available at this
seat this round.
