# Red round 1 — lens 3 (leaf-node citation verification, slice 3 of 4: §4 + §5)

Slice rule applied: report sections §0–§8 divided evenly across 4 instances; slice 3 = §4 (sharded
findings + collator) and §5 (round-scoped audit), including their footnotes. Full report re-read in
context before auditing (914 lines, two windowed reads). Ledger was empty at pass start — every
pair below is first-verification this run.

## Verdict (lens-scoped, slice only)

Slice PASSES at the substance level: every load-bearing §4/§5 claim traced to its source and
corroborated HIGH, including all first-hand re-verifications (pin equivalence, file sizes, engine
source lines, friction entries, ledger lines 159–187, R4-4/R5-1/R5-2 blocks, backlog 28(d),
retro rows 10/18/23, PromptCaching rates, ContextLength figures). Five LOW / LOW-MEDIUM
citation-hygiene gaps found — none verdict-relevant, all trivial fixes. No HIGH or MEDIUM-HIGH
findings in this slice.

## Findings

### L3-F1 — footnote misattributes its quoted phrase to the wrong pinned file — LOW
certain (text defect, mechanically confirmed) × low (substance holds across sources) × trivial.
**Location:** §5.1, "5 incomplete-propagation chains in 5 rounds (...; plus R4-4's unpropagated
fifth numeral), each costing a full audit round at $25–30.[^PropagationChains]" and the footnote
"[^PropagationChains]: `report.md` §3 row 23 corrected enumeration + §2.1(b) @ `bfa8a3b`: ...
'5 chains in 5 rounds.'"
**Problem:** the quoted phrase "5 chains in 5 rounds" does not appear in the retrospective
report.md — grep returns zero hits; that report says "four chains in this corpus" (line 1541).
The 5-chain phrasing lives in `blue-researcher.md` line 14 ("5 chains in run 3's 5 rounds; each
one bought a whole extra audit round" — which also carries the per-round cost claim) and
`debate.js` line 263 ("5 regressions in 5 rounds"). The chain enumeration itself IS verbatim in
report.md rows 23/§2.1(b) (verified, lines 633/908), and 4 chains + R4-4's numeral = 5 is sound
arithmetic — but the footnote attributes a quotation to a file that contradicts its count.
**Required fix:** re-source the quoted phrase and the "each costing a full audit round" claim to
`blue-researcher.md` line 14 @ `5396952` (or drop the quotation marks); keep report.md for the
enumeration only.

### L3-F2 — [^Backlog28d] lever list misquotes the source's lever (4) — LOW
certain × low (the correctly-cited part carries §4's argument; lever 4 is uncited elsewhere) ×
trivial.
**Location:** Footnotes, "[^Backlog28d]: ... levers (1) shard, (2) collator, (3) prompt-level
read batching, (4) per-agent timeline."
**Problem:** backlog item 28(d) @ `5396952` reads "(4) TOOLING step-up if gap volume grows:
evaluate beads ... vs a tiny sc-gaps Go tool"; the per-agent timeline is the item's trailing
sentence ("cost.md should show a per-agent timeline"), not lever (4). The retrospective's own
footnote (report.md line 1093) quotes the lever list correctly — the error is new in this run's
report. §8 open question 2's use ("backlog 28(d) names it") survives — the timeline IS named,
just not as lever 4.
**Required fix:** "(4) tooling step-up (beads vs sc-gaps); plus a per-agent-timeline note".

### L3-F3 — Votta prose attributes the replication's false-positive result to Votta — LOW-MEDIUM
medium (a reader citing onward will mis-cite) × low (the argumentative point — consolidation's
value is judgment/false-positive filtering — is corroborated by the literature bundle either
way) × trivial.
**Location:** §4.6 item 5, "Votta found collection meetings added few defects over independent
reviews BUT were significantly better at false-positive reduction.[^Votta]"
**Problem:** the minimal-synergy result (individuals already record ~9 of 10 meeting defects) is
Votta 1993. The *significant false-positive reduction* result traces to the replication line —
the Springer Empirical Software Engineering study the footnote itself parenthetically names
("Does Every Inspection Really Need a Meeting?", link.springer.com/article/10.1023/A:1009787822215)
and successors — not to Votta's own paper. ACM primary returned 403 (paywall); corroboration via
search-level sources only — the false-positive attribution could not be pinned to Votta's
abstract and the replication record points away from him. Pattern: footnote over-attribution /
within-source condition misattribution.
**Required fix:** "Votta found meetings added few defects; the EMSE replication found them
significantly better at false-positive reduction" — one clause split.

### L3-F4 — "12.5× cache-write" repeats a pinned artifact's internal units-vs-ratio error — LOW
certain (recomputable from the source's own pricing header) × low (the judgment-tier-premium
conclusion survives at the correct 5×) × trivial.
**Location:** §4.2 first bullet, "rate-driven at the judgment tier (5× cache-read, 12.5×
cache-write).[^CostAudit]"
**Problem:** cost.md's own pricing line (sonnet cache-write 2.5 $/MTok; session model 12.5
$/MTok) makes the cache-write multiplier 5×, same as cache-read; 12.5 is the absolute $/MTok
rate, not a ratio. Blue faithfully transcribes cost.md finding 3's own error — notable because
§6.4.1 demonstrates blue audited the SAME artifact's finding 2 as internally contradicted while
inheriting finding 3's defect unchecked. Belongs beside finding 2 in the §6.4 defect list for
the artifact's successor.
**Required fix:** "(5× cache-read, 5× cache-write — the session tier's 12.5 $/MTok vs sonnet's
2.5)"; add to §6.4 as cost.md defect (b).

### L3-F5 — PR #18 cited as shipped without discharging the input's own verify-merged order — LOW
low (verified true first-hand this pass: PR #18 merged at `4a3801c`, an ancestor of the pin
`5396952`) × low × trivial.
**Location:** §5.3, "PR #18's recall layer (lex + vec + hyde retrieval, hook-refreshed on every
markdown write) is a paraphrase-tolerant site-finder...[^AlreadyShipped]" against the footnote's
own text "PR #18 (qmd recall layer — open at staging time; verify merged before citing as shipped)."
**Problem:** the report leans on PR #18 as available capability while its cited source still
carries the unresolved verify-before-citing instruction and no verification is recorded anywhere
in the report. Red closed the loop: `git log` shows merge commit `4a3801c` (feat/qmd-recall-layer)
precedes the pin, and the hook/`.mcp.json` artifacts exist in the working tree — so the claim is
TRUE; only the record is incomplete. Note the interaction with this run's friction.md entry 1
(seats spawned before `.mcp.json` existed had no qmd tools), which §8 Q9 already carries.
**Required fix:** one footnote clause: "verified merged (`4a3801c`, pre-pin) at round 1."

## Verification log (statement ↔ reference, graded)

| # | Claim (§4/§5) | Verified against | Confidence |
|---|---|---|---|
| 1 | Full-re-read MUST names blue's living report; no own-archive re-read mandate in agent file or either red prompt | red-auditor.md l.13/l.20 + debate.js l.212/216, direct read | HIGH |
| 2 | Merge $7.52/13.22/12.64/10.60/13.56 Σ$57.54 (38%); red-lens $9.22–11.05 Σ$49.48 (excl. r6 $0.61); blue-respond Σ$18.21; r4+r5 Σ$53.00 | cost.md table, recomputed | HIGH |
| 3 | cost.md finding 2 contradicted by own table (r5 $13.56>$13.22, 7.87M>5.64M reads, 61 turns, 6-gap board) | cost.md + findings.md round-5 block | HIGH |
| 4 | findings.md 106,772 B / 1364 lines; blue/report.md 159,394 B; pin diffs empty | wc + git diff first-hand | HIGH |
| 5 | Friction #4 (filename-keyed guard, scratchpad control), #6 (enum rounding, verbatim), #15 (25k cap names blue/report.md), #11 (skip-rule held), #8 (r3 precautionary detours) | friction.md direct read | HIGH |
| 6 | Friction #10 → "ledger appended via cat across four rounds, zero incidents" | entry attests r4 positively; other rounds by ledger blocks + absence of contrary friction | MEDIUM-HIGH |
| 7 | Ledger l.159 first MA-status entry (R5-2 never ledgered); l.184 six-id overrule; l.185 lens-5 hold overruled via report lines 496/727 | citation-ledger.md direct read | HIGH |
| 8 | R5-1 block quote (three lenses 1/2/4; chain links vs own closure entries); R4-4 l.503 grep quote; R5-2 block (cross-corpus drift, source sentence mid-round-2, lens-3 direct read of other corpus) | findings.md direct read | HIGH (R5-2 "unchanged since round 2" MEDIUM-HIGH — inferred from "written mid-round-2") |
| 9 | debate.js: no-fs comment, citationPasses l.198, ledgerClause drift triggers l.205, lens prompt l.212, "genuinely new gaps only" merge prompt, lineage throw, whole-debate window l.186/244–245, propagation sentence l.263, degenerate-FAIL guard, row-16b header | direct read @ working tree = pin | HIGH |
| 10 | blue-researcher.md l.14 propagation clause | direct read | HIGH |
| 11 | Backlog 28(d): TURNS x CONTEXT, ~100–150K / 2.7M+ reads, levers 1–3 verbatim | ideas/backlog.md direct read | HIGH (lever-4 misquote = L3-F2) |
| 12 | Retro rows 10 (R2-9 ledger repair), 18 (hold + sharding first candidate), 23 + §2.1(b) enumerations | retro report.md direct read | HIGH |
| 13 | §5.2 catch descriptions (R4-4 by grep at merge; R3-6 zero-hits grep class; R3-10 direct read) | findings.md l.503/20/24/44 | HIGH for R4-4; MEDIUM-HIGH for R3-6/R3-10 (class-level match, original blocks not re-read) |
| 14 | PR #18 merged pre-pin; sc-recall-index hook + .mcp.json qmd entry exist | git log + grep first-hand | HIGH |
| 15 | This-run friction #8: blue-synthesize Write of blue/report.md refused, neutral-name+cp detour | friction.md (this run) direct read | HIGH |
| 16 | [^ContextLength] 13.9%–85% even at perfect retrieval; whitespace/masked conditions | arXiv:2510.05381 abstract, leaf-fetched | HIGH |
| 17 | [^PromptCaching] 0.1× read / 1.25× 5-min / 2× 1-hr; prefix billed at read rate per call | platform.claude.com live doc, leaf-fetched | HIGH (volatile: pricing page) |
| 18 | [^Votta] minimal synergy | search-level corroboration (ACM 403) | MEDIUM (attribution split = L3-F3) |
| 19 | [^SafeRTS]/[^YooHarman]/[^LostMiddle]/[^FentonOhlsson] definitional claims | established literature, consistent with stated results; primaries paywalled/not re-fetched | MEDIUM-HIGH |
| 20 | [^HandoffLoss]/[^HierSumm]/[^DiffReview] shape-only claims | as-labeled by the report (volatility noted, non-load-bearing); not fetched | MEDIUM (as-labeled — acceptable) |

## Friction
- ACM-hosted primaries (Votta 10.1145/167049.167070, SafeRTS 10.1145/248233.248262) return 403 to
  WebFetch and no PDF MCP tools were exposed at this lens seat (no ToolSearch/MCP available) —
  the protocol's pdf-reader MUST-try clause was unsatisfiable here; Votta graded via search-level
  corroboration instead of the primary.
- qmd MCP retrieval likewise not exposed at this seat; local Grep/Read sufficed for the corpus
  slice (mode-2/mode-3 unaffected).
