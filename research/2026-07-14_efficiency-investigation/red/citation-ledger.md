# red citation-ledger — Efficiency and termination levers for the frank-exchange-of-views debate engine

## Round 1 — lens 4 (slice: §6–§8 + carried footnotes)

cost.md per-seat table & findings (merge $57.54/38%, respond $18.21/12%, r5 merge $13.56/7.87M/61t vs r2 $13.22/5.64M, finding-2 text, ~15 mostly-trivial, stop-resume) | research/2026-07-12_feov-retrospective/cost.md @ bfa8a3b | HIGH | R1 | 2026-07-14
Backlog severity-floor spec + "would have ended run 3 at round 3 for ~$10" + conceded-error quote | ideas/backlog.md item 30 @ 5396952 | HIGH | R1 | 2026-07-14
Backlog 28(d) turns×context + red-merge-r1 ~100-150K/2.7M+ quote | ideas/backlog.md item 28 @ 5396952 | HIGH | R1 | 2026-07-14 (footnote's lever-(4) enumeration wrong → L4-F3)
Round-3 board: 10 open, 2 MEDIUM-HIGH (R3-1, R3-2), both code-trace, complexity low | red/findings.md @ bfa8a3b lines 620-621/717/737 | HIGH | R1 | 2026-07-14
R4-1 grading quote (certain × high, four of five lenses) | red/findings.md line 425 | HIGH | R1 | 2026-07-14
R4-4 grep quote ("4th|fourth", exactly one uncorrected) | red/findings.md line 503 | HIGH | R1 | 2026-07-14
debate.md headers: 11 anchored (6 BLUE/5 RED/0 LEAD); unanchored "### LEAD" = 5 | live grep re-run at this seat | HIGH | R1 | 2026-07-14
friction #15 names blue/report.md (frontier H4 says findings.md — misattribution real) | retrospective friction.md entry 15 + blue/frontier.md H4 | HIGH | R1 | 2026-07-14
frontier H5 rides on R4-3 as type specimen (blue's §5.1 correction accurate) | blue/frontier.md H5 + findings.md R4-3 line 485 | HIGH | R1 | 2026-07-14
friction #12 Grep count-mode footgun (64 vs 66) | retrospective friction.md entry 12 | HIGH | R1 | 2026-07-14
This-run friction: memory unreadable ×3 seats; synthesis write-block detour; count-mode recurrence | efficiency-investigation friction.md | HIGH | R1 | 2026-07-14
Lane-2 disconfirming budget 4/13 (31%) | blue/candidates/lane-2.md lines 9, 416 | HIGH | R1 | 2026-07-14
File sizes: findings.md 106,772 B / 1364 lines; blue/report.md 159,394 B | wc at pin-equal working tree | HIGH | R1 | 2026-07-14
Pin equivalence: both git diff --stat empty (bfa8a3b, 5396952) | re-run live | HIGH | R1 | 2026-07-14
CVSS 68% inconsistent ratings, n=59 | arXiv:2308.15259 abstract, leaf-fetched | HIGH | R1 | 2026-07-14
~34% NVD-vs-CNA disagreement | arXiv:2508.13644 abs + /html/v1, leaf-fetched | LOW — FIGURE ABSENT from cited paper (misattribution) → L4-F1 | R1 | 2026-07-14

[round-1 lens-3, slice §4/§5] red-auditor.md l.13 full-re-read MUST verbatim; no own-archive re-read mandate in agent file or red prompts (debate.js l.212/216) | direct read at working tree = pin | HIGH | round 1 | 2026-07-14
[round-1 lens-3] cost.md: merge $7.52/13.22/12.64/10.60/13.56 Σ$57.54=38%; red-lens $9.22–11.05 Σ$49.48; blue-respond Σ$18.21; r4+r5 Σ$53.00; finding-2 self-contradiction (r5 $13.56>r2 $13.22, 7.87M>5.64M, 61 turns); finding 5 verbatim | cost.md, recomputed | HIGH | round 1 | 2026-07-14
[round-1 lens-3, DEFECT] §4.2 "12.5× cache-write" — cost.md's own pricing header makes the ratio 5× (2.5 vs 12.5 $/MTok); 12.5 is absolute rate, not multiplier | recompute from cost.md header | HIGH (L3-F4) | round 1 | 2026-07-14
[round-1 lens-3] findings.md 106,772 B / 1364 lines; blue/report.md 159,394 B; pin diffs empty (bfa8a3b, 5396952) | wc + git diff first-hand | HIGH | round 1 | 2026-07-14
[round-1 lens-3] friction #4 (filename-keyed guard isolation), #6 (enum-rounding quote verbatim), #15 (25k cap names blue/report.md), #11 (skip-rule held all confidences) | run-3 friction.md direct read | HIGH | round 1 | 2026-07-14
[round-1 lens-3] "ledger appended via cat across four rounds, zero incidents" | friction #10 attests r4 positively; rest inferred from ledger blocks + no contrary friction | MEDIUM-HIGH | round 1 | 2026-07-14
[round-1 lens-3] run-3 ledger l.159 = first MA-status entry (R5-2 never a ledgered pair); l.184 six-id overrule; l.185 lens-5 hold overruled via report lines 496/727 | run-3 citation-ledger.md direct read | HIGH | round 1 | 2026-07-14
[round-1 lens-3] R5-1 block (lenses 1/2/4, chain links vs own closures), R4-4 l.503 grep quote, R5-2 block (cross-corpus drift; source sentence written mid-round-2) | run-3 findings.md direct read | HIGH (R5-2 "unchanged since round 2" MEDIUM-HIGH) | round 1 | 2026-07-14
[round-1 lens-3] debate.js: no-fs comment l.32–34, citationPasses l.198, ledger drift clause l.205, lens prompt l.212 (FULL living report = blue/report.md), merge prompt "genuinely new gaps only", lineage throw l.227–235, whole-debate window l.186/244–245, blue propagation sentence l.263, degenerate-FAIL guard, row-16b KNOWN TRADEOFF header | direct read | HIGH | round 1 | 2026-07-14
[round-1 lens-3] blue-researcher.md l.14 propagation clause ("5 chains in run 3's 5 rounds; each one bought a whole extra audit round") | direct read | HIGH | round 1 | 2026-07-14
[round-1 lens-3, DEFECT] [^PropagationChains] quoted phrase "5 chains in 5 rounds" NOT in retro report.md (says "four chains", l.1541); phrase lives in blue-researcher.md l.14 / debate.js l.263; enumeration itself verbatim at retro l.633/908 | grep + direct read | HIGH (L3-F1) | round 1 | 2026-07-14
[round-1 lens-3] backlog 28(d): TURNS x CONTEXT, ~100–150K / 2.7M+ cache reads, levers (1) shard (2) collator (3) prompt-level read batching verbatim | ideas/backlog.md direct read | HIGH | round 1 | 2026-07-14
[round-1 lens-3, DEFECT] [^Backlog28d] "(4) per-agent timeline" — source's lever (4) is tooling step-up (beads vs sc-gaps); timeline is a trailing sentence | ideas/backlog.md + retro report l.1093 | HIGH (L3-F2) | round 1 | 2026-07-14
[round-1 lens-3] retro report rows 10 (R2-9 ledger repair), 18 (hold + sharding as first candidate scoping rule), 23 + §2.1(b) corrected enumerations | retro report.md direct read | HIGH | round 1 | 2026-07-14
[round-1 lens-3] §5.2 catch-to-arm rows: R4-4 caught by grep at merge (findings l.503 exact); R3-6/R3-10 class-level match (findings l.20/24/44) | run-3 findings.md | HIGH (R4-4) / MEDIUM-HIGH (R3-6, R3-10) | round 1 | 2026-07-14
[round-1 lens-3] PR #18 merged pre-pin (merge commit 4a3801c ancestor of 5396952); sc-recall-index hook + .mcp.json qmd entries present | git log + grep first-hand | HIGH (record gap = L3-F5) | round 1 | 2026-07-14
[round-1 lens-3] this-run friction #8: blue-synthesize Write of blue/report.md refused, neutral-name+cp detour (§4.3(c) live-confirmation claim) | this run's friction.md direct read | HIGH | round 1 | 2026-07-14
[round-1 lens-3] [^ContextLength] "13.9%–85%" degradation at perfect retrieval; whitespace/masked conditions confirmed in abstract | https://arxiv.org/abs/2510.05381 leaf-fetched | HIGH | round 1 | 2026-07-14
[round-1 lens-3] [^PromptCaching] 0.1× cache-read / 1.25× 5-min write / 2× 1-hr write; cached prefix billed at read rate per call | platform.claude.com prompt-caching doc leaf-fetched (volatile: pricing) | HIGH | round 1 | 2026-07-14
[round-1 lens-3] [^Votta] minimal meeting synergy = Votta 1993; significant false-positive reduction = Springer EMSE replication line, not Votta primary (ACM 403, no PDF MCP at seat) | WebSearch corroboration | MEDIUM (L3-F3) | round 1 | 2026-07-14
[round-1 lens-3] [^SafeRTS]/[^YooHarman]/[^LostMiddle]/[^FentonOhlsson] definitional claims consistent with established results; primaries paywalled, not re-fetched | knowledge-level check | MEDIUM-HIGH | round 1 | 2026-07-14
[round-1 lens-3] [^HandoffLoss]/[^HierSumm]/[^DiffReview] used shape-only, volatility self-labeled, non-load-bearing per report | as-labeled, not fetched | MEDIUM | round 1 | 2026-07-14

## Round 1 — lens 1 (slice 1: preamble + §0 + §1)

Pin equivalence (both git diff --stat empty) | [^PinCheck] | HIGH (re-run first-hand) | 1 | 2026-07-14
Severity-floor spec + "ended run 3 at round 3 for ~$10" quote | [^BacklogLevers] backlog.md:30 @5396952 | HIGH | 1 | 2026-07-14
log()-heartbeat item still open | [^BacklogLevers] — actually backlog.md:31 (adjacent item; see L1-F3) | MEDIUM (substance true, attribution wrong) | 1 | 2026-07-14
"conceded an error" quote in docket-detector item | [^BacklogLevers] backlog.md:29 | HIGH | 1 | 2026-07-14
Board table: 20/11/10/5/6 open; max HIGH/MH/MH/HIGH/MH; members as named | [^FindingsBoard] findings.md 1281,1287,1080,1112,1156,1170,1181,717,737,425,200 | HIGH | 1 | 2026-07-14
R3-1/R3-2 both MEDIUM-HIGH, code-trace, complexity low | [^FindingsBoard] findings.md:717,737 | HIGH | 1 | 2026-07-14
Round-3 RED "severity declining... 2 MEDIUM-HIGH, both code-trace" | [^Round3Red] debate.md:491-492 (round-3 RED section) | HIGH | 1 | 2026-07-14
R4-1 grading quote + "four of five lenses" | [^R4OneDetail] findings.md:425 | HIGH | 1 | 2026-07-14
R5-5 grading quote (telemetry-invisible) | [^R5FiveDetail] findings.md:200 | HIGH | 1 | 2026-07-14
cost.md per-seat figures; Σ57.54/38%; Σ49.48/33%; Σ18.21; rounds 4-5 = $53.00 recomputed | [^CostAudit] cost.md table | HIGH | 1 | 2026-07-14
Stop-and-resume ~$0, ~7 rounds cut, round-6 lenses killed | [^StopResume] cost.md finding 5 | HIGH | 1 | 2026-07-14
FAIL / UNVERIFIED / 6 open at ceiling | findings.md:77 + run-3 report.md:3 | HIGH | 1 | 2026-07-14
Ceiling disposition table (grading/response/disposition/rationale) | [^CeilingDisposition] run-3 report.md:7-20 | HIGH | 1 | 2026-07-14
PR #15 core = lineage docket + enforcement throw (R4-1/R5-5 fixes) | [^AlreadyShipped] inputs/already-shipped.md | HIGH | 1 | 2026-07-14
KS < 0.05 for 2 consecutive rounds stopping criterion | [^AdaptiveStability] arxiv.org/html/2510.12697v1 (leaf-fetched) | HIGH (body claim); MEDIUM (footnote's "rounds 2-8, <1%" gloss — see L1-F2) | 1 | 2026-07-14
Debate saturation ~2-5 rounds, decline after 2 on some tasks | [^DebateRounds] arxiv.org/html/2506.00066v1 (leaf-fetched) | HIGH | 1 | 2026-07-14
Dalal-Mallows stop rule keyed to observed count vs cost ratio; JASA 83(403):872-879 | [^DalalMallows] via Höhle 2016 exposition (primary 403-paywalled) | HIGH (via detailed secondary) | 1 | 2026-07-14
STADS residual risk from discovery curve; Good-Turing singleton | [^Stads] arxiv abs/1803.02130 + search digest (FSE21 PDF lossy; no PDF MCP at seat) | MEDIUM-HIGH | 1 | 2026-07-14
Briand quote "no model is sufficiently accurate and underestimation may be substantial" | [^CaptureRecaptureEval] ieeexplore 852741 abstract (quote-matched) | HIGH | 1 | 2026-07-14
Petersson JSS 2004 estimator-bias / localization-mismatch claims | [^CaptureRecaptureDecade] wohlin.eu/jss04-1.pdf (lossy fetch; bibliographic only) | MEDIUM | 1 | 2026-07-14

[round-1 lens 2] pin equivalence re-verified: both `git diff --stat` checks empty (retrospective @ bfa8a3b; backlog+plugin @ 5396952) | git, first-hand | HIGH | round 1 | 2026-07-14
[round-1 lens 2] backlog item 30 carries verbatim: severity-floor "~$10" claim, risk-mass umbrella + never-zero spot-check caveat, grade_disputes schema, best-of-N "lone-voice bias survives" precondition | ideas/backlog.md item 30 @ 5396952 | HIGH | round 1 | 2026-07-14
[round-1 lens 2] debate.js: citationPasses recomputed per round `Math.min(4, Math.max(1, Math.ceil((blueEnv.claim_count || 20) / 40)))`; judge enum exactly closed|rebuttal_sustained|risk_accepted|carried|unresolved (no grade-wrong value); contested window = whole debate via allPriorGapIds + supersedes-descendant; lineage throw present | debate.js @ 5396952, direct read | HIGH | round 1 | 2026-07-14
[round-1 lens 2] run-3 debate.md: anchored ^### grep = 11 headers (6 BLUE, 5 RED, 0 LEAD); unanchored "### LEAD" count = 5 (footgun reproduces) | debate.md @ bfa8a3b, re-run first-hand | HIGH | round 1 | 2026-07-14
[round-1 lens 2] R4-1 header (HIGH certain x high x low-medium, "four of five lenses"), R5-5 header (MEDIUM-HIGH medium x high, telemetry-invisible), R5-1 header (certain x medium, "lenses 1, 2, 4") all verbatim | red/findings.md @ bfa8a3b lines 425/200/135 | HIGH | round 1 | 2026-07-14
[round-1 lens 2] R5-5 singleton: lens-5 enforcement argument present; lens-4 "Considered raising this... Not raised" at lines 115-120; lens-2 supersedes mentions detector-mechanics only | round-5 lens files @ bfa8a3b | HIGH | round 1 | 2026-07-14
[round-1 lens 2] R4-1 minted by lenses 1/2/3/5 (lens-2 "Finding 1 (NEW, round 4) — the contested-docket detector is lineage-blind by construction" verbatim line 19; lens-4 docket hits unrelated) | round-4 lens files @ bfa8a3b | HIGH | round 1 | 2026-07-14
[round-1 lens 2] run-3 citation-ledger lines ~184-185: six-id overrule (five ids by extraction) and no-discrepancy-hold overrule (report lines 496+727) both present as report describes | red/citation-ledger.md @ bfa8a3b | HIGH | round 1 | 2026-07-14
[round-1 lens 2] cost.md: red-lens sum r1-5 = $49.48 exactly (33% of $149.95); blue-respond 3.95/3.96/2.98/3.05/4.27 exact; red-merge 7.52/13.22/12.64/10.60/13.56 exact; "rounds 3-5 closed ~15 mostly-trivial gaps" verbatim | cost.md @ bfa8a3b | HIGH | round 1 | 2026-07-14
[round-1 lens 2] round-5 mass recomputed under disclosed mapping = 31.0 exactly (7+6+3.5+3.5+6+5) — brackets lane values ~31/~32; "round 5 rose" holds | red/findings.md round-5 grade headers, recomputed first-hand | HIGH (round 5; other rounds on two-lane agreement) | round 1 | 2026-07-14
[round-1 lens 2] round-4 BLUE "No rebuttals this round — every gap was real..." verbatim; round-5 BLUE concedes/builds all six R5-1..R5-6, zero grade contests; round-2 RED risk-accept negotiation quote verbatim | debate.md @ bfa8a3b | HIGH | round 1 | 2026-07-14
[round-1 lens 2] run-3 report rows 15/23: R2-1 count 3->2 with High retained by 2-for-2 argument; R3-7 narrowed, grade kept; R5-1 enumeration corrected, grades untouched | report.md §3 @ bfa8a3b | HIGH | round 1 | 2026-07-14
[round-1 lens 2] NineJudges: all four quoted results verbatim in abstract (2 votes / 8-22 pts / best single >= panel / <=11%) | arXiv:2605.29800 | HIGH | round 1 | 2026-07-14
[round-1 lens 2] PoLL: disjoint-family smaller-model panel beats single large judge, less intra-model bias, >7x cheaper — verbatim | arXiv:2404.18796 | HIGH | round 1 | 2026-07-14
[round-1 lens 2] CVSS inconsistency: "59 participants... 68% gave different severity ratings" verbatim | arXiv:2308.15259 | HIGH | round 1 | 2026-07-14
[round-1 lens 2] persuasive debate: 76%/88% vs naive 48%/60% verbatim | arXiv:2402.06782 | HIGH | round 1 | 2026-07-14
[round-1 lens 2] WeakJudges: debate beats consultancy in ALL scenarios; task-dependence is vs direct QA; "more modestly" = stronger-debater gains — footnote gloss misassigns condition (L2-F2) | arXiv:2407.04622 | HIGH source / footnote defect | round 1 | 2026-07-14
[round-1 lens 2] sequential-test savings: cited paper says 36-75% reduction, report claims "30-50% typical" (L2-F6) | arXiv:2603.00216 | MEDIUM (band mismatch) | round 1 | 2026-07-14
[round-1 lens 2] few-reviewers underestimation substance ("at least four to five reviewers... acceptable (Briand et al., 2000)"; Mh-JK non-robust at 2, underestimates) | Petersson JSS 2004 PDF via wohlin.eu + pdftotext | HIGH substance; Briand verbatim quote MEDIUM (IEEE blank page) | round 1 | 2026-07-14
[round-1 lens 2] Iso29119 taxonomy companion exists, title/topic verbatim; 29119 normative gloss itself unfetched | arXiv:1905.10676 | HIGH title / MEDIUM gloss | round 1 | 2026-07-14
[round-1 lens 2] ConflictingScores: full HTML contains NO NVD-vs-CNA rate and NO ~34% figure (inter-system comparison only); paper is open-access, NOT paywalled as §7 claims — misattribution (L2-F1) | arXiv:2508.13644 /html/v1 | CONTRADICTED | round 1 | 2026-07-14
[round-1 lens 2] frontier H3 text and red-auditor.md line 13 full-re-read MUST both match report characterizations | blue/frontier.md; agents/red-auditor.md @ 5396952 | HIGH | round 1 | 2026-07-14
[round-1 lens 2] ExpertCvss (paywalled), RbtTaxonomy, DalalMallows, FentonOhlsson accepted as-labeled (blue self-graded MEDIUM where digest-sourced; classical characterizations match known literature) | — | MEDIUM / MEDIUM-HIGH | round 1 | 2026-07-14

## Round 1 — merge re-verifications (red-merge-r1)

debate.js control flow: PASS break l.236 precedes contested block l.244-245; carried never enters adjudicated (l.252-253, append-only); judge prompt l.249 reads debate.md + red/findings.md in full; findings.md path hardcoded at l.216 AND l.249 | debate.js @ pin-equal working tree, direct read | HIGH | R1 | 2026-07-14
Round-2 MEDIUM-HIGH members R2-1/R2-3/R2-7/R2-8/R2-9 all carry complexity "low" (R1-4's premise: no floor fires after r3 but not r2) | run-3 red/findings.md ll.1080/1112/1156/1170/1181 @ bfa8a3b, direct read | HIGH | R1 | 2026-07-14
Blue §5.2 catch-to-arm table rows = R4-3/R4-4/R3-6/R3-10/R5-1/R1-1 — R5-2 absent despite §5.1 naming it a clean specimen (R1-3) | blue/report.md §5.1-§5.2 side-by-side read | HIGH | R1 | 2026-07-14
Frontier H1 grades R5-5 "HIGH" and R4-1 "High likelihood x High impact" vs pinned MEDIUM-HIGH / certain x high (R1-19) | blue/frontier.md H1 + run-3 findings.md ll.200/425 | HIGH | R1 | 2026-07-14

## Round 2 — lens 4 (slice: §6–§8 + footnotes)

Pin equivalence re-run: both git diff --stat empty | git first-hand | HIGH | R2 | 2026-07-14
[^CathedralBazaar] 44,180 Pairwise / 72,122 Consumer-View / 194/266=73% / 139/288=48% / AC-UI-Impact / "accuracy can drop by 40%" | arXiv:2607.05670 abs + /html/v1, leaf-fetched twice | HIGH | R2 | 2026-07-14
[^CostAudit] r1 additions: r6 killed spawn $0.61 (table row); "31 gaps ($60-ish)" + "~15 mostly-trivial" verbatim; header 2.5 vs 12.5 → 5× multiplier; red-lens Σ$49.48 recomputed; 38.4/33.0/12.1% recomputed | cost.md @ bfa8a3b direct read + recompute | HIGH | R2 | 2026-07-14
§6.4 item 5: "54KB" verbatim in run-3 friction #15 AND backlog 31(g); matches neither pinned size | friction.md e15 + backlog.md l.31 direct read | HIGH | R2 | 2026-07-14
[^BacklogLevers] heartbeat = backlog l.31 STILL OPEN list ("NEXT PR: emit log() at every state transition") | backlog.md l.31 direct read | HIGH | R2 | 2026-07-14
[^Sprt] "by at least 36% and by at most 75%" exact sentence present; condition = SYMMETRIC ERROR BOUNDS — footnote/§2.3 drop the condition → L4-F2 | arXiv:2603.00216 leaf-fetched | HIGH on numbers / MEDIUM as written | R2 | 2026-07-14
[^AdaptiveStability] −1.03% delta (BIG-Bench/Gemma) + KS<0.05×2 criterion + 10-round baseline confirmed; "stops ~4–7" CONTRADICTED — Table 2 spans 2–8 (JudgeAnything 2,2,8); round-0 "2–8" matched the table's min–max → L4-F1 repair-regression (vector: red R1-20 phrasing) | arXiv:2510.12697v1 leaf-fetched twice | LOW as written (band) / HIGH (criterion+delta) | R2 | 2026-07-14
§7 "paywalled ONLY [^ExpertCvss]+Fragmentation" vs in-report Votta ACM 403 + DalalMallows 403-at-red-seat (via-secondary provenance undisclosed in footnote) → L4-F3 | report-internal cross-read + R1 ledger l.63 | MEDIUM (disclosure defect, not corroboration failure) | R2 | 2026-07-14
Repairs faithful to red's R1 facts, no re-fetch needed: [^PropagationChains]/[^Backlog28d]/[^Votta]/[^WeakJudges]/[^FrontierH3]/[^ConflictingScores]/[^AlreadyShipped] | ledger cross-check vs edited footnote text | HIGH/MEDIUM as ledgered | R2 | 2026-07-14

## Round 2 — lens 2 (slice 2: §2–§3)

[round-2 lens-2] CathedralBazaar arXiv:2607.05670: 44,180 Pairwise / 72,122 Consumer-View; "194/266 (73%) and 139/288 (48%) of CNAs have a median of at least 1 vector divergent" verbatim (PDF §IV/§V-A; arXiv-HTML garbles it to "at least 11" — italic artifact, PDF authoritative); AC/UI/Impact concentration + 40% transfer drop in abstract | arXiv abs + /html/v1 + PDF via pdftotext, leaf-fetched | HIGH | round 2 | 2026-07-14
[round-2 lens-2] Sprt arXiv:2603.00216: "the sequential test reduces the average sample size by at least 36% and by at most 75%" — band faithful; SOURCE CONDITION "for symmetric error bounds" dropped by §2.3's "at matched error rates" gloss (L2-F3); quoted fragment also drops second "by" | arXiv abs, leaf-fetched | HIGH figure / condition gap noted | round 2 | 2026-07-14
[round-2 lens-2] Fragmentation of CVSS scores (C&S 2026, S0167404826001549): 403 reproduced first-hand at this seat — §2.2/§7 "genuinely paywalled, remains unverified" corroborated | WebFetch 403 | HIGH (access-failure claim) | round 2 | 2026-07-14
[round-2 lens-2] §2.5 item 1 "cost.md produced post-run by scripts/cost-audit.mjs from token metering, which never sees grades" — TRUE: script parses harness per-agent transcripts' usage blocks, zero grade/journal references | cost-audit.mjs direct read | HIGH | round 2 | 2026-07-14
[round-2 lens-2, DEFECT] §2.5 item 1 sink chain "log() line into trajectories/journal.jsonl, consumed by cost-audit.mjs" — CONTRADICTED both halves: run-3 journal.jsonl holds only started/result events (0 log() lines: grep "researching:|debate ended" = 0); cost-audit.mjs contains 0 journal.jsonl references (input = workflow transcript dir) | run-3 journal.jsonl + cost-audit.mjs first-hand (L2-F1) | HIGH (contradiction) | round 2 | 2026-07-14
[round-2 lens-2] §3.5/§0 row 3a tally: lane-2 conditional vote explicitly discounted at both sites — conditional-vote-laundering class does NOT recur | report direct read both sites | HIGH | round 2 | 2026-07-14
[round-2 lens-2] §2.4 arithmetic: 49.48+0.61=50.09 ✓; 12/149.95=8.0% ✓; ~$2/lens × 2 lenses × 3 rounds ≈ $12 ✓ | recomputed | HIGH | round 2 | 2026-07-14
[round-2 lens-2] §3 unchanged-claim spot-set (zero LEAD headers; $7.52–$13.56 range; PASS-break l.236 < contested l.244; judge enum lacks grade-wrong; R3-2 three-rounds-unnoticed; blue-respond series) all HIGH in round-1 ledger, sections' round-1 edits consistent with pinned entries | ledger + CHANGELOG cross-read | HIGH (carried) | round 2 | 2026-07-14

## Round 2 — lens 1 (slice 1: preamble + §0 + §1 + §2)

[round-2 lens-1] CathedralBazaar independently re-verified (44,180 / 72,122 / 194/266=73% / 139/288=48% / AC-UI-Impact / 40% drop / Zhang-Massacci-Zhang title+authors) | arXiv:2607.05670 abs + /html/v1, third independent fetch — agrees with lens-4 l.102 and lens-2 l.113 | HIGH | R2 | 2026-07-14
[round-2 lens-1] AdaptiveStability criterion "Dt<0.05 for 2 consecutive rounds" verbatim + largest delta -1.03% confirmed; my fetch enumerated only BIG-Bench/JudgeBench rows (4-7), prose hinted 4-8 — DEFER to lens-4 l.107's twice-fetched JudgeAnything 2,2,8: "~4-7" band contradicted; conflict surfaced to merge, not held | arXiv /html/2510.12697v1 | HIGH (criterion+delta) / LOW as written (band, per lens-4) | R2 | 2026-07-14
[round-2 lens-1] §1.2 ~$78 counterfactual (r3 = 2.98+9.46+12.64 = $25.08; r4-5 = $53.00) | cost.md recomputed | HIGH | R2 | 2026-07-14
[round-2 lens-1] §1.2 "no threshold setting reproduces the round-3 stop" — sound within severity×fix-cost family (r2 and r3 boards identical on both axes per ledgered grades) | logic on ledger l.95 premises | HIGH | R2 | 2026-07-14
[round-2 lens-1] §1.5 live-code correction semantics (PASS break l.236 → contested l.244-245 → same-round judge dispatch l.247-250; carried excluded from adjudicated l.252-253; ids accumulate l.258; degenerate-FAIL guard l.225) | debate.js direct read | HIGH | R2 | 2026-07-14
[round-2 lens-1] §2.4 figures ($0.61 r6 spawn; $50.09 seat total; ~$4/round x 3 = $12/run ≈ 8% of $149.95) | cost.md recomputed — agrees with lens-2 l.119 | HIGH | R2 | 2026-07-14
[round-2 lens-1] §2.3 R4-1 "all three citation instances" (run-3 round-4 lenses 1/2/3 = leaf-node citation verification per file headers; lens 5 = dark-side) | run-3 red/candidates headers direct read | HIGH | R2 | 2026-07-14
[round-2 lens-1, DEFECT] §2.5 item 1 sink+consumer: run-3 journal.jsonl = 87 lines, only started/result dispatch records, zero log() output (l.52/l.271 signatures absent); cost-audit.mjs reads only agent-*.jsonl usage | first-hand, independent of lens-2 l.117 — convergent (L1-F1) | HIGH (contradiction) | R2 | 2026-07-14
[round-2 lens-1] blue-researcher.md l.14 "5 chains in run 3's 5 rounds" verbatim; debate.js l.263 reads "5 regressions in 5 rounds" — [^PropagationChains] names both as homes of the quoted phrase; only one is | direct read | HIGH (half-misquote, handed to §5/footnotes instance) | R2 | 2026-07-14
[round-2 lens-1] §2.2 "genuinely paywalled (403 at the abstract)" — 403 reproduced (lens-2 l.115) corroborates access failure; the word "genuinely paywalled" still asserts a mechanism a 403 cannot show (L1-F3, LOW) | reasoning + lens-2 corroboration | MEDIUM | R2 | 2026-07-14

## Round 2 — lens 3 (slice: §4/§5 + carried footnotes)

[round-2 lens-3] pin equivalence re-run: both git diff --stat empty (bfa8a3b, 5396952) | git first-hand | HIGH | round 2 | 2026-07-14
[round-2 lens-3] §4.2 R1-30 repair: cost.md header sonnet cr 0.2/cw 2.5, session cr 1/cw 12.5 → both multipliers 5×; finding-3 l.35 carries "12.5x cache-write" internally | cost.md ll.3/35 direct read | HIGH | round 2 | 2026-07-14
[round-2 lens-3] §4.3(c) R1-34 repair: "54KB" verbatim at run-3 friction.md l.21 (entry 15) AND backlog.md l.31 item (g) | direct read both sites | HIGH | round 2 | 2026-07-14
[round-2 lens-3] §4.5 cond-7 base rate: "R3-2's schema-declared friction field uncalled for three rounds" verbatim | run-3 findings.md l.200 (R5-5 header) | HIGH | round 2 | 2026-07-14
[round-2 lens-3] §4.6 item-1 sixth-file hazard: merge prompt directory-scoped ("lens passes in red/candidates/") | debate.js l.216 region direct read | HIGH | round 2 | 2026-07-14
[round-2 lens-3] §5.2 R5-2 row: cross-corpus drift caught by direct read of MA corpus findings; source sentence written mid-round-2 | run-3 findings.md ll.157-175 | HIGH ("unchanged since round 2" stays MEDIUM-HIGH) | round 2 | 2026-07-14
[round-2 lens-3] §5.2 R4-3 row: "caught independently by lenses 2 and 4", same cell as R3-5's correction | run-3 findings.md l.485 | HIGH | round 2 | 2026-07-14
[round-2 lens-3] §5.5 gate cond-3: "this clause outranks any token saving" verbatim in research-protocol mode-2 clause | live skill text | HIGH | round 2 | 2026-07-14
[round-2 lens-3] §5.3 PR#18 hook-refresh claim: hooks.json PostToolUse Write|Edit → sc-recall-index; main.go .md gate → qmd update | code direct read | HIGH | round 2 | 2026-07-14
[round-2 lens-3, DEFECT L3-F1] §4.2 "run-3 transcripts gitignored and absent at the pin"/unavailable — agent-transcripts.tar.gz PRESENT in working tree: 7,040,514 B, 46 per-agent jsonl, mtime 02:19 predates run launch (PINNED.md 06:02) | ls + tar -tzf + mtimes first-hand | CONTRADICTED (working tree) | round 2 | 2026-07-14
[round-2 lens-3, DEFECT L3-F2] §4.6 item-1 mechanism: cat redirect is Bash, never fires the Write|Edit hook; double-index path is the next qmd update's collection sweep | hooks.json + sc-recall-index main.go | HIGH (defect minor, conclusion intact) | round 2 | 2026-07-14
[round-2 lens-3] propagation sweep slice tokens (54KB / 12.5 / smallest): all surviving hits inside correction records | grep report-wide | HIGH | round 2 | 2026-07-14

## Round 2 — merge block (first-hand re-verifications at the merge seat)

[round-2 merge] run-3 journal.jsonl = 87 lines (46 started + 41 result), zero log() signatures; cost-audit.mjs zero journal.jsonl references | wc/grep first-hand — confirms L1-F1/L2-F1 (R2-1) | CONTRADICTS §2.5 item 1 sink/consumer | round 2 | 2026-07-14
[round-2 merge] run-3 agent-transcripts.tar.gz present: 7,040,514 B, 46 members, mtime 02:19 < PINNED.md 06:02 | ls + tar -tzf first-hand — confirms L3-F1 (R2-3) | CONTRADICTS §4.2 unavailability premise | round 2 | 2026-07-14
[round-2 merge] citationPasses = min(4, max(1, ceil(claim_count/40))) + 2 fixed lenses; rounds 1–2 candidate dirs hold 6 lens files each | debate.js + ls first-hand (R2-2) | HIGH (mechanism) | round 2 | 2026-07-14
[round-2 merge] debate.js: no fs usage; PASS break l.236 / deadlock l.256 / while l.192; one judge dispatch per round ll.247–250; judge full-read prompt l.249 | direct read (R2-4/R2-5/R2-6/R2-11) | HIGH | round 2 | 2026-07-14
[round-2 merge] blue-researcher.md l.14 "5 chains in run 3's 5 rounds" verbatim; debate.js l.263 "5 regressions in 5 rounds" | direct read — confirms L1 handoff (R2-15), overrules L3's fidelity-based HIGH | HIGH (half-misquote) | round 2 | 2026-07-14

## Round 3 — lens 2 (slice 2: §2–§3)

[round-3 lens-2] [^Sprt] R2-14 repair re-verified as a new claim: "for symmetric error bounds, the sequential test reduces the average sample size by at least 36\% and by at most 75\%" verbatim in abstract — quotation full, condition faithful, no repair-regression | arXiv:2603.00216 abs re-fetched | HIGH | round 3 | 2026-07-14
[round-3 lens-2] [^JournalCheck] (R2-1 repair): debate.js `log(`researching: ...`)` confirmed at l.52 region (direct read); LIVE journal wf_5cefd2a4-35f re-checked at round 3 = 50 lines (28 started + 22 result), zero "researching"/log() output — the console-ephemeral claim still holds as the journal grows; cost-audit.mjs direct read: input filter startsWith('agent-') && endsWith('.jsonl') (l.28 region), zero journal.jsonl references | first-hand | HIGH | round 3 | 2026-07-14
[round-3 lens-2] §2.4/§7 "this run's rounds 1 and 2 both demonstrably ran 6 [lens seats]" — red/candidates holds round-{1,2}-lens-{1..6}.md, 12 files exactly | ls first-hand | HIGH | round 3 | 2026-07-14
[round-3 lens-2] §2.4 R2-2 arithmetic: 3 agents × ~$2/lens × 3 rounds = ~$18; ~$18/~$180 rescaled baseline ≈ 10% ✓ (3-round basis unstated in-section — L2-F2) | recomputed | HIGH (arithmetic) | round 3 | 2026-07-14
[round-3 lens-2] §2.2 R2-18/R2-12 repairs textually faithful ("403 shows access failure, not mechanism"; throttle-input subject restored); "the journal is subscription" hedged-plausible (Elsevier Computers & Security) | direct read + knowledge-level | HIGH (repairs) / MEDIUM-HIGH (hedge) | round 3 | 2026-07-14
[round-3 lens-2] Carried HIGH, no round-2 edits touching them: §3.3(vi) exits ll.192/236/256 (R2 merge l.155); run-3 journal 87=46+41 (R2 merge l.152); §2.2 68%/73%/44,180; §2.1 mass series; §3.1 zero-LEAD; §3.6 NineJudges/PoLL/merge-range $7.52–$13.56; §3.5 tally non-laundered (R2 l.118) | ledger cross-check vs CHANGELOG | HIGH (carried) | round 3 | 2026-07-14

## Round 3 — lens 4 (slice: §6–§8 + footnotes)

[round-3 lens-4] [^Sprt] restored full quotation "for symmetric error bounds, the sequential test reduces the average sample size by at least 36% and by at most 75%" — EXACT word-for-word match; R2-14 repair clean | arXiv:2603.00216 abstract leaf-fetched | HIGH | round 3 | 2026-07-14
[round-3 lens-4] [^AdaptiveStability] amended gloss fully re-verified as a new claim: Table 2 = 22 configurations (multimodal MLLM-Judge + JudgeAnything run 3 vision models each, not 4 — a naive 6×4=24 count is wrong; row enumeration totals 4+4+4+4+3+3=22); range 2–8; exactly 3 rows outside 4–7 (JudgeAnything 2, 2, 8 — both ends); remaining 19 all in 4–7 ("typically 4–7" holds); max delta −1.03% (Gemma-3-4B/BIG-Bench); fixed 10-round baseline | arXiv:2510.12697v1 leaf-fetched twice (second fetch row-enumerated) | HIGH — R2-13 repair clean | round 3 | 2026-07-14
[round-3 lens-4] [^JournalCheck] claim (a) live-journal shape re-verified first-hand at this seat: journal now 50 lines = 28 started + 22 result, grep 'researching|debate ended' = 0, zero non-lifecycle lines (footnote's 43/22/21 was a mid-run snapshot; counts grew, shape identical) | wf_5cefd2a4-35f/journal.jsonl direct read | HIGH (shape claim; counts volatile as self-labeled) | round 3 | 2026-07-14
[round-3 lens-4] [^MergeDecomposition] "identified by cost-audit.mjs's own 'Red merge, round N' head match" — regex /Red merge, round (\d+)/ present at cost-audit.mjs l.33 | direct read | HIGH | round 3 | 2026-07-14
[round-3 lens-4, GAP] [^MergeDecomposition] reproducibility: the ~70-line parser script is NOT preserved anywhere in the run dir (glob *.mjs/*.js = zero hits) and its input tarball is untracked — §4.2's table, §6.1 item-1's ranking and §8 Q2's ANSWERED all rest on blue self-report re-derivable only from method prose (L4-F1) | Glob + footnote text | MEDIUM-HIGH (figures as blue-measured; preservation gap LOW-MEDIUM) | round 3 | 2026-07-14
[round-3 lens-4] [^DalalMallows] via-secondary access note matches red ledger l.63 verbatim (Höhle 2016, HIGH-via-detailed-secondary); [^PropagationChains] round-2 amendment matches merge first-hand reads (ledger l.156) incl. the lens-overrule parenthetical | ledger cross-check | HIGH | round 3 | 2026-07-14
[round-3 lens-4] §7 "four access-blocked sources" enumeration: no fifth attempted-and-blocked source exists report-wide (ExpertCvss/Fragmentation/Votta/DalalMallows are the only 403-or-paywall-attempted set); SafeRTS/YooHarman/LostMiddle/FentonOhlsson primaries remain UNATTEMPTED, accepted round 1 at knowledge-level MEDIUM-HIGH (ledger l.42) — carried acceptance, not access-blocked, no new gap | report-wide cross-read + R1 ledger | HIGH (enumeration) | round 3 | 2026-07-14
[round-3 lens-4] §6.1 rescale arithmetic: $49.48/25 lens-rounds ≈ $1.98/lens → ~+$2/round, ~+$10/run over 5 rounds ✓; ceil(160/40)=4+2=6 ✓ — live-corroborated: this round dispatched exactly 4 citation instances + 2 fixed lenses | recompute + this round's own dispatch shape | HIGH | round 3 | 2026-07-14

## Round 3 — lens 1 (slice 1: preamble + §0 + §1 + §2)

[round-3 lens-1] pin equivalence re-run: both git diff --stat empty (bfa8a3b, 5396952) | git first-hand | HIGH | R3 | 2026-07-14
[round-3 lens-1] AdaptiveStability amended gloss (R2-13): Table 2 = 22 rows by hand-count of enumerated rows (fetch model's own summary mis-said 18 — enumeration trusted, summary not); 3 outside 4-7 = JudgeAnything 2,2,8 both ends; typically-4-7 fair (19/22); max delta -1.03% BIG-Bench/Gemma; Dt<0.05 x2 criterion | arXiv /html/2510.12697v1 re-fetched | HIGH | R3 | 2026-07-14
[round-3 lens-1] Sprt restored quotation (R2-14): "for symmetric error bounds, the sequential test reduces the average sample size by at least 36% and by at most 75%" verbatim in abstract, condition prefix same sentence, second "by" present | arXiv abs/2603.00216 re-fetched | HIGH | R3 | 2026-07-14
[round-3 lens-1] JournalCheck clause (a) live journal: wf_5cefd2a4-35f/journal.jsonl = 50 lines (28 started + 22 result) at this seat, zero other types, "researching" grep 0 — blue's 43 (22+21) consistent with growth; composition claim exact | first-hand | HIGH | R3 | 2026-07-14
[round-3 lens-1] §1.2 R2-10 arithmetic: 78-10=68 net; 10+5+6=21 mints; series total 52 = 31+15+6 vs cost.md closure counts | recomputed | HIGH | R3 | 2026-07-14
[round-3 lens-1] §2.4 R2-2 rescale: ~$2/lens x 3 agents x 3 rounds = $18; lens-candidate ingest 52-80KB/round recomputed from §4.2 table (80/69/52.5/62.7/60.4) | recomputed | HIGH | R3 | 2026-07-14
[round-3 lens-1] §2.2 R2-12 subject restored + R2-18 adverb corrected, both present by direct read; §0 rows 1-2 sink conditionality propagated; row 3a conditional-vote discount intact | report direct read | HIGH | R3 | 2026-07-14
[round-3 lens-1, NOTE L1-F1 LOW] §2.1 item 4 "mean mass per gap stays ~5" — recomputed means 4.9/5.9/4.4/6.0/5.2 (band 4.4-6.0, ~20% stretch at r4); no-downward-trend conclusion holds, precision nit only | recomputed from report's own tables | MEDIUM-HIGH as written | R3 | 2026-07-14

## Round 3 — lens 3 (slice 3: §4/§5 + carried footnotes)

[round-3 lens-3] §4.2 decomposition table (5 rounds: totals 174/247/250/190/318KB; shares incl. r4 blue 20/findings 32/candidates 33; findings 0.1%→29-32%; 60-92KB r4-5) | agent-transcripts.tar.gz extracted, 5 red-merge transcripts recomputed first-hand (raw + cache-weighted) | HIGH | round 3 | 2026-07-14
[round-3 lens-3] §4.2 cache-weighted findings dollars ~$1.40/2.60/4.10/4.10, Σ≈$12/run findings-attributable | recomputed: $1.44/2.64/4.18/4.16, Σ$12.4 (bytes × remaining turns, share × merge $) | HIGH | round 3 | 2026-07-14
[round-3 lens-3] §4.2 "145KB ingested at round 5" (blue/report.md) | recomputed 146.9KB | HIGH | round 3 | 2026-07-14
[round-3 lens-3, DEFECT L3-F1] §4.2 + §8 Q2 "blue/report.md is the largest merge-context component every round from round 2" | own table r4 row + recompute: r4 blue 20.0% < candidates 33.6% < findings 31.8% (cache-weighted blue 14.9%) — false at round 4 | CONTRADICTED | round 3 | 2026-07-14
[round-3 lens-3, DEFECT L3-F2] §4.2 "landing INSIDE lane 1's 20-30K directional estimate" | lane-1.md l.281: estimate is the ARCHIVE's share by round 5; measured 15-23K is whole-file r4-5; r4 15K below band; archive sub-fraction ≤23K at/below floor | HIGH (mismatch first-hand) | round 3 | 2026-07-14
[round-3 lens-3] §4.2 "$7-10 squarely in lane 1's $5-15 band" | lane-1.md l.283 "~$5-15 at run-3 scale" — subset holds | HIGH | round 3 | 2026-07-14
[round-3 lens-3, DEFECT L3-F3] §4.2/§6.1 "sharding-addressable ≈ $7-10/run" via archive-fraction "clear majority" | [^MergeDecomposition] method is per-FILE attribution; no documented step splits findings.md into archive/open — sub-figure not reproducible from stated method | LOW-MEDIUM (assertion; $12 parent figure HIGH) | round 3 | 2026-07-14
[round-3 lens-3] §4.5 cond 6 quotes "A stub world" / "canned envelopes and no live agents" | tests/simulator/harness.mjs ll.24/7 direct read | HIGH | round 3 | 2026-07-14
[round-3 lens-3] [^MergeDecomposition] "identified by cost-audit.mjs's own 'Red merge, round N' head match" | cost-audit.mjs l.33 regex verbatim | HIGH | round 3 | 2026-07-14
[round-3 lens-3] tarball 46 members, extractable | tar -xzf first-hand, 46 agent-*.jsonl | HIGH | round 3 | 2026-07-14
[round-3 lens-3] §4.6 item 2 "52-80KB/round" lens-candidate merge ingest | recomputed 55-81KB (matches table's rounded cells) | HIGH | round 3 | 2026-07-14
[round-3 lens-3] R2-4/R2-5/R2-16 repairs (§4.5 cond 7/2, §4.6 item 1) | ledgered HIGH rounds 1-2 (debate.js ll.216/227-235/244-245/249; hooks.json) — carried, spot-checked consistent | HIGH (carried) | round 3 | 2026-07-14

## Round 3 — merge block (first-hand at the merge seat)

[round-3 merge] JUDGE_ENVELOPE schema: required ['deadlock','resolutions']; resolution items gap_id|resolution|rationale only; enum incl. carried; judge dispatch fires AFTER redEnv submitted (ll.244-258) — no judge-writable demanded-read field exists | debate.js ll.125-144, 244-258 direct read | HIGH | R3 | 2026-07-14
[round-3 merge] R3-1(b): zero *.mjs/*.js anywhere under the run directory (find first-hand); retrospective trajectories/ still holds agent-transcripts.tar.gz + journal.jsonl (untracked-present) | filesystem first-hand | HIGH | R3 | 2026-07-14
[round-3 merge] R3-5: lane-1.md l.281 reads "the archive's share of merge context is maybe 20-30K tokens by round 5" — archive SHARE, round 5; §4.2 compares whole-file r4-5 | lane-1.md direct read | HIGH on the mismatch | R3 | 2026-07-14
[round-3 merge] R3-13: §4.2 table row sums recomputed 86.1 / 80 / 91 / 91 / 96% — residual 4-20%/round undisclosed | report's own printed cells | HIGH | R3 | 2026-07-14
[round-3 merge] R3-4: §4.2 round-4 row (blue 20% < findings 32% < candidates 33%) vs "largest every round from round 2" sentence, same section; §8 Q2 carries the same sentence | report direct read + L3's transcript recompute | HIGH (contradiction) | R3 | 2026-07-14
[round-3 merge] R3-9(b) vector: red's own R2-11 finding text reads "clause (v)'s new dispute-resolution enum value" — blue copied it; the enum lives in §3.3's resolution-enum bullet, not clause (v) | prior findings.md + report side by side | HIGH | R3 | 2026-07-14
[round-3 merge] R3-6: §1.5 arm (b) "two consecutive rounds minted zero new gaps above the floor" vs registered prediction "a final round whose ... new-mint list is empty" — single-round, total-mint trigger does not test the hardened arm | report §1.5 side-by-side read | HIGH | R3 | 2026-07-14

## Round 4 — lens 2 (slice 2: §2–§3)

[round-4 lens-2] pin equivalence re-run: both git diff --stat empty (bfa8a3b, 5396952) | git first-hand | HIGH | round 4 | 2026-07-14
[round-4 lens-2] NineJudges all four results verbatim ("about 2 independent votes' worth" / "8-22 percentage points short" / "best single judge matches or outperforms the full panel across all conditions" / "close at most 11% of this gap") | arXiv:2605.29800 abs re-fetched (>2 rounds since R1) | HIGH | round 4 | 2026-07-14
[round-4 lens-2] PoLL: smaller-model disjoint-family panel outperforms single large judge, "over seven times less expensive", "less intra-model bias" | arXiv:2404.18796 abs re-fetched | HIGH | round 4 | 2026-07-14
[round-4 lens-2] PersuasiveDebate: "76% and 88% accuracy respectively (naive baselines obtain 48% and 60%)" verbatim | arXiv:2402.06782 abs re-fetched | HIGH | round 4 | 2026-07-14
[round-4 lens-2] WeakJudges: consultancy beaten "across all tasks" (random-assignment condition); vs direct QA "mixed" without info asymmetry; stronger-debater "more modestly than in previous studies" — R1-22 gloss still faithful | arXiv:2407.04622 abs re-fetched | HIGH | round 4 | 2026-07-14
[round-4 lens-2] CvssInconsistent: "59 participants ... 68% of these users gave different severity ratings" verbatim | arXiv:2308.15259 abs re-fetched | HIGH | round 4 | 2026-07-14
[round-4 lens-2] Iso29119 taxonomy title + 29119 coverage confirmed; normative gloss still unfetched (standard itself) | arXiv:1905.10676 abs re-fetched | HIGH title / MEDIUM gloss | round 4 | 2026-07-14
[round-4 lens-2] Briand quote "no model is sufficiently accurate and underestimation may be substantial": ieeexplore rendered empty at this seat; exact phrase corroborated via search record of same paper (ResearchGate) — R1 leaf quote-match stands | search corroboration | HIGH (carried R1 + R4 secondary) | round 4 | 2026-07-14
[round-4 lens-2] §2.1 item 4 (R3-14): means 4.9/5.9/4.4/6.0/5.2 recomputed from printed lane-1 row; band 4.4-6.0; rounds 2+4 highest; certain×low=3.5 vs medium×medium=4 | recomputed | HIGH | round 4 | 2026-07-14
[round-4 lens-2] R3-7/R3-8/R3-10 §2 repairs present and faithful (presence-check honesty + mass re-derivation clause; mapping pinned/new-series/realized-excluded consistent across §2.5/§2.1/§8-Q6; ×3-rounds basis stated, $6×3=$18, ~10% of ~$180) | report direct read + recompute | HIGH | round 4 | 2026-07-14
[round-4 lens-2] R3-3 heads (b),(c) faithful ("NOT equivalent"; cumulative threshold + overflow batch-docket); R3-12 terminal set + pricing faithful; R3-9 cross-ref consistent with one-dispatch-per-round ll.247-250 | report + debate.js direct read | HIGH | round 4 | 2026-07-14
[round-4 lens-2, DEFECT L2-F1] §3.3(v) "a git-tracked surface every seat and the human operator already read" — CONTRADICTED at the pin: lens prompt reads report.md/CHANGELOG/ledger only (never debate.md; live-corroborated by this dispatch's own prompt); red-merge prompt appends ### RED with no read mandate; blue reads latest ### RED; judge reads debate.md in full only when a docket fires. Red-side seats — the adversarial watchmen — have no read path; "any seat (blue, a lens, ...) may docket" names a lens without read path or judge channel. Chain R1-2→R2-6→R3-3→this | debate.js ll.195-265 direct read | HIGH (contradiction) | round 4 | 2026-07-14
[round-4 lens-2] judge enum (closed|rebuttal_sustained|risk_accepted|carried|unresolved, no grade-wrong) + citationPasses min(4,max(1,ceil(cc/40))) re-read first-hand | debate.js | HIGH | round 4 | 2026-07-14

## Round 4 — lens 1 (slice 1: preamble + §0 + §1 + §2)

[round-4 lens-1] pin equivalence re-run: both git diff --stat empty (bfa8a3b, 5396952) | git first-hand | HIGH | R4 | 2026-07-14
[round-4 lens-1] §1.1/§2 grade quotes re-read at pinned lines: R5-1 l.135, R5-5 l.200, R4-1 l.425 (certain×high, four of five lenses), R3-1/R3-2 ll.717/737 both MEDIUM-HIGH code-trace complexity low, R2-1/3/7/8/9 ll.1080-1181 all MEDIUM-HIGH complexity low; backlog item 30 at l.30 | run-3 findings.md + backlog.md direct read | HIGH | R4 | 2026-07-14
[round-4 lens-1] §1.5 "round-2 closure ledger records both as closed WITH REGRESSION" — this run's findings.md ll.11/335/412 record the round-3 amendment reclassifying R1-12/R1-17 closed-clean → closed-with-regression | red/findings.md direct read | HIGH | R4 | 2026-07-14
[round-4 lens-1] §1.5 R3-6 restated prediction composes with hardened arm (i↔b; ii stricter than a — TRUE implies arm fired; iii↔c; single-round trigger disavowed; $25-30 − $10 ≈ $15-20) | report-internal re-derivation | HIGH | R4 | 2026-07-14
[round-4 lens-1] §2.1 R3-14 band recomputed exactly: 4.9/5.9/4.4/6.0/5.2; two highest r2/r4; certain×low=3.5 vs medium×medium=4 faithful to mapping | recompute from printed tables | HIGH | R4 | 2026-07-14
[round-4 lens-1] §2.5 R3-7/R3-8 repairs consistent with §8 Q6 at both sites (realized excluded; mapping enumerated; new-series rule; non-red-merge recompute executor named) | report direct read | HIGH | R4 | 2026-07-14
[round-4 lens-1, DEFECT L1-F1] §2.4 "3 throttled rounds = low-mass rounds 3-5": throttle input for round 3 is the post-round-2 board (~65, run's second-highest); §2.1 item 1 names only post-r3/post-r4 as lowest — stated basis reads own-round post-mass (lookahead); honest band $12-18/run; vector: ×3 originated in red's R2-2 rescale | §2.1 vs §2.4 report-internal | CONTRADICTED as printed (timing) | R4 | 2026-07-14
[round-4 lens-1] [^DebateRounds] re-fetched: saturation quotes at 3/4/5 rounds; "only until the second round, after which accuracy declines" — §1.3 gloss and footnote both hold | arXiv /html/2506.00066v1 | HIGH | R4 | 2026-07-14
[round-4 lens-1] [^CvssInconsistent] re-fetched: 59 participants, "68% of these users gave different severity ratings" verbatim | arXiv:2308.15259 abs | HIGH | R4 | 2026-07-14
[round-4 lens-1] [^Stads] re-fetched: residual-risk-from-discovery + ecological-biostatistics confirmed at abstract; Good-Turing specifics body-level | arXiv:1803.02130 abs | MEDIUM-HIGH (as labeled) | R4 | 2026-07-14
[round-4 lens-1] [^RbtTaxonomy] arXiv:1912.11519 leaf-resolved FIRST time: title/authors exact (Felderer & Schieferdecker, "A taxonomy of risk-based testing") | arXiv abs | HIGH identity / MEDIUM gloss (self-labeled) | R4 | 2026-07-14

## Round 4 — lens 4 (slice: §6–§8 + footnotes)

[round-4 lens-4] decompose-merge.mjs RE-RUN first-hand vs fresh tarball extract (46 members): raw totals 172.7/245.8/249.2/188.2/316.6KB; strict findings $ series 0.26/0.53/0.89/1.16 EXACT (Σ$2.84); r5 blue 145.7KB / findings 28.8%/91.2KB / 61 turns; residual 13.6/19.5/9.0/8.6/4.0% exact; r4 row 20.1/31.7/33.6 | instrument re-run at this seat | HIGH — R3-1 figures reproduce | R4 | 2026-07-14
[round-4 lens-4] 72.6% archive fraction: awk re-run = 28867/76356/105223 (72.57%); findings.md l.340 verbatim "Verdict (round 4): FAIL — superseded by round 5, preserved" | first-hand | HIGH | R4 | 2026-07-14
[round-4 lens-4, DEFECT L4-F1] "committed as trajectories/decompose-merge.mjs" — CONTRADICTED at leaf: git status `?? trajectories/` (whole dir untracked, ditto blue/ and red/candidates/), git log --all for the file empty, not gitignored — present in working tree, absent from object store; lead's R3-1(b) said "commit the parser into the run dir" | git first-hand | figures HIGH / status claim FALSE (LOW-MEDIUM gap) | R4 | 2026-07-14
[round-4 lens-4, DEFECT L4-F2] "the only .gitignore entry" — repo .gitignore has ~12 entries; tarball pattern at l.21 is only the sole entry AFFECTING run trajectories | .gitignore direct read | LOW (false universal) | R4 | 2026-07-14
[round-4 lens-4] proportional-share ceiling: Σ(share × cache-read$) = 0.00+0.50+0.91+1.74+2.27 ≈ $5.4/run vs stated "≈$6"; $3–6 band and $2–4 composition (0.726 × 2.8–6) hold | cost.md merge rows + measured shares, recomputed | HIGH (band) / MEDIUM-HIGH ("≈$6") | R4 | 2026-07-14
[round-4 lens-4] §8 Q6 DECIDED text matches lead R3-8 ruling ("this run is the owner of record", debate.md ll.572–577); mapping = §2.1's with realized excluded | debate.md + report direct read | HIGH | R4 | 2026-07-14
[round-4 lens-4] §6.4 item 6 enum pointer (R3-9 repair): clause (v) carries no enum; resolution-enum bullet separate in §3.3 | report direct read | HIGH | R4 | 2026-07-14
[round-4 lens-4] §7 claim-count echo: ceil(166/40)=5 → capped 4 + 2 fixed = 6 seats; live-corroborated (this pass = citation instance 4 of 4) | arithmetic + dispatch shape | HIGH | R4 | 2026-07-14
[round-4 lens-4] cost.md defects re-confirmed: l.35 "12.5x cache-write" vs header 2.5/12.5 → 5×; finding-2 dispute-size sentence vs its own r5 row | cost.md direct read | HIGH (carried) | R4 | 2026-07-14
[round-4 lens-4, NOTE L4-F3] §6.1 "comparable to item 2's batching saving" — §4.6 item 2's own $1–2/round ≈ $5–10/run EXCEEDS sharding's $2–4/run; rank argued on non-dollar grounds, phrasing understates | report-internal arithmetic | LOW | R4 | 2026-07-14

## Round 4 — lens 3 (slice 3: §4/§5 + footnotes)

[round-4 lens-3] pin equivalence re-run: both git diff --stat empty (bfa8a3b, 5396952) | git first-hand | HIGH | round 4 | 2026-07-14
[round-4 lens-3] §4.2/[^MergeDecomposition] raw decomposition table + totals 173/246/249/188/317KB + r5 blue 145.7KB + 61/62 turns vs cost.md + strict $0.26/0.53/0.89/1.16 Σ$2.84 + "other" 13.6/19.5/9.0/8.6/4.0 + column sums 86/80/91/91/96 | decompose-merge.mjs RE-RUN first-hand against pinned tarball at this seat | HIGH (all reproduce) | round 4 | 2026-07-14
[round-4 lens-3] §4.2 archive fraction: awk recompute 76,356/105,223 = 72.57%; l.340 boundary verbatim "Verdict (round 4): FAIL — superseded by round 5, preserved" | pinned findings.md first-hand | HIGH | round 4 | 2026-07-14
[round-4 lens-3] §4.2 ceiling cross-checks: r5 cache-read 7.87M/$7.87; 1.16/7.87≈15%; 4.10/7.87=52%; proportional ceiling recomputed from cost.md cache-read column: raw-share Σ≈$5.4, cw-share Σ≈$6.3 — "≈$6" within rounding | cost.md rows + recompute | HIGH | round 4 | 2026-07-14
[round-4 lens-3] R3-4 repair: largest r2/r3/r5 (36.3/43.0/46.0), r4 candidates 33.6>findings 31.7>blue 20.1 — faithful at §4.2 + §8 Q2 | instrument re-run | HIGH | round 4 | 2026-07-14
[round-4 lens-3] R3-5 repair: r5 findings whole-file 91.2KB≈22.8K tok; 72.6%×23K≈16.6K below 20–30K floor | instrument re-run + lane-1.md ledgered | HIGH | round 4 | 2026-07-14
[round-4 lens-3] §4.5 cond-7 judge home: round-3 ### LEAD entry exists (debate.md l.481), rulings' rationales name records read, git-tracked — "demonstrates the form by construction" | debate.md direct read | HIGH | round 4 | 2026-07-14
[round-4 lens-3] [^R4FourGrep] re-verified after 3 rounds: findings.md l.503 verbatim | grep first-hand | HIGH | round 4 | 2026-07-14
[round-4 lens-3, DEFECT L3-F1] §4.2/[^MergeDecomposition]/§6.1/§8 Q2 "committed as trajectories/decompose-merge.mjs" — file PRESENT (4,228 B, runs, not gitignored) but UNTRACKED: git status ??, zero commits, run-dir tracked set = skeleton only | git first-hand | CONTRADICTED (present ≠ committed) | round 4 | 2026-07-14
[round-4 lens-3, DEFECT L3-F2] §4.2 "reproduces only if cache-weighted BYTES are priced as tokens" — strict×4 = $1.04/2.12/3.56/4.64 (Σ$11.4) ≠ printed ~$1.40/2.60/4.10/4.10; cw-share × whole-merge-$ reproduces both printed and lens-3 series to ≤3% ($1.41/2.64/4.16/4.14 Σ$12.35); ledger l.192 records lens-3 method as "share × merge $" | instrument re-run + cost.md + own ledger | CONTRADICTED (mechanism; corrected $3–6 band itself reproduces, HIGH) | round 4 | 2026-07-14
[round-4 lens-3, NIT L3-F3] [^MergeDecomposition] "the only .gitignore entry" — .gitignore holds ~14 entries; tarball line is the only trajectories-scoped one | .gitignore first-hand | LOW (imprecision) | round 4 | 2026-07-14

## Round 4 — merge block

[round-4 merge] "committed as trajectories/decompose-merge.mjs" | git re-run at merge seat: `?? trajectories/`, ls-files empty for dir, `git log --all -- '*decompose-merge*'` empty; file present 4,228 B | CONTRADICTED (present, untracked) — R4-2 | R4 | 2026-07-14
[round-4 merge] R4-1 arithmetic, both directions | strict×4 = $1.04/2.12/3.56/4.64 vs printed ~$1.40/2.60/4.10/4.10 (off 15–35%/round); printed ÷ cost.md merge $ = 10.6/20.6/38.7/30.2% shares, matching instrument's measured cache-weighted shares ≤0.5pt — printed series IS share × merge $ | HIGH (recomputed at merge) | R4 | 2026-07-14
[round-4 merge] lens-3 round-3 method characterization in §4.2/[^MergeDecomposition] | this ledger's own l.192: "(bytes × remaining turns, share × merge $)" — NOT bytes-priced-as-tokens | CONTRADICTED (misattribution) — R4-1 | R4 | 2026-07-14
[round-4 merge] §3.3(v) "a git-tracked surface every seat ... already read" | debate.js l.212 re-read at merge: lens prompt names blue/report.md + CHANGELOG (+ ledger); no debate.md; judge reads it only when a contested docket fires (ll.244–250) | CONTRADICTED for lens seats — R4-3 | R4 | 2026-07-14
[round-4 merge] GRADE enum has 8 members incl. `trivial`; Q6 pin maps 7 | debate.js l.60 direct read at merge | HIGH — R4-5 | R4 | 2026-07-14
[round-4 merge] [^MergeDecomposition] "the only .gitignore entry" | repo .gitignore direct read at merge: 14 entries; tarball line (l.21) is the only trajectories-scoped one | CONTRADICTED as universal — R4-10 | R4 | 2026-07-14
[round-4 merge] run-record capture manifest names no instrument-class file | research-protocol SKILL.md run-dir tree: trajectories/ = "journal.jsonl (tracked) + agent-transcripts.tar.gz (gitignored)" only | HIGH — R4-2's convention-not-mechanism head | R4 | 2026-07-14
[round-4 merge] §2.4 "low-mass rounds 3–5" vs §2.1 series | mass table 98/65/44/30/31 re-read at merge; post-round-2 (~65) = second-highest; §2.1 item 1 names only post-r3/post-r4 as lowest predecessors | HIGH — R4-7 | R4 | 2026-07-14
[round-4 merge] §4.5 condition 6: skeleton-born headline vs red-merge opening-act addition | both clauses re-read side by side at merge — two creators named for the same files | HIGH (report-internal) — R4-6 | R4 | 2026-07-14
[round-4 merge, AMENDMENT to l.192] the round-3 HIGH on lens 3's Σ$12.4 recompute stands for what the entry records (a faithful recompute under the share × merge $ convention, reproducing the printed series) — but the convention prices findings' share of WHOLE-merge dollars (cache-write + output + overhead), not cache-read attribution; the $3–6 band supersedes it as the findings-attributable figure; re-examined per blue's round-3 request | own ledger + instrument + cost.md | amended, disclosed | R4 | 2026-07-14
