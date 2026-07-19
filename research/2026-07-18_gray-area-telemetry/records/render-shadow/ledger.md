# red/ledger.md — RENDERED PROJECTION (source of truth: records/ event log; do not hand-edit)

## OPEN GAPS (9)

### R1-9 — Three lens findings converge on one hole: the report states the recommendation's limit but never arg
S8: "it is still self-report, so it buys durability and non-circularity, not sincerity."
severity low | low-medium x low-medium | cx low-medium | class unargued-case | found_by L5-F2,L5-F3,L6-F5
regraded x1 (history in the event log; latest basis: Blue shipped the primary leg — the parity answer and the disconfirmability argument at §8. The residue is one unstated cost grade (adjudication-time verification cost, graded cheaper/equal/higher), not an unargued case. Impact unchanged at low-medium.)
required_fix: Add one paragraph to S8 answering: given both channels are self-report, why durability and non-circularity are worth the switch, and what artifact verification costs at adjudication time relative to reading thinking blocks. Conceding parity is an acceptable answer; silence is not.
acceptance_check: DOCUMENT-PROBE: read S8 and confirm an explicit answer to the parity objection and an explicit statement of verification cost (cheaper, equal, or higher) is present.

### R2-1 — R1-2 required blue to state whether the 287/5,569 figures derive from the pinned probe and if so ret
§2 "Empirical state of the local store" — "Lane-3's independent earlier sweep reported 287 transcripts and 5,569 blocks with the same all-empty result."
severity low-medium | medium x low-medium | cx trivial | class incomplete-repair-lag | supersedes R1-2
required_fix: Determine whether lane-3 sweep and inputs/probe-thinking-persistence.md are the same measurement (both recoverable), then evidence independence at §2 or strike the word there. CLASS RULE: every provisional retraction filed in Provenance must propagate to every site asserting the retracted property; sweep the retracted token in both directions. Enumeration open.
acceptance_check: DOCUMENT-PROBE: grep -n "independent" blue/report.md; no line may assert independence of the 287/5,569 sweep while another retracts it. If asserted, a named comparison of the two sources must appear.

### R2-2 — The guard blue quotes (mbc) fires only on isNonInteractiveSession. On the interactive branch with sh
Catechism answer 3(b) — "This is a causal finding: the resolver guard directly produces the observed empty blocks." And §2 — "The all-empty local sweep is therefore the expected output of a default-configured install, not evidence of a defect."
severity medium | medium-high x medium | cx trivial | class causal-overreach | supersedes R1-7 R1-2 | found_by L6-F2,L6-F16
regraded x1 (history in the event log; latest basis: Two of three acceptance-check legs shipped in round 2 (Catechism 3(b) and §2 both partition the causal claim by session type and concede the interactive mechanism unresolved). The surviving defect is one un-carved-out clause in the Provenance limitations section while both prominent sites carry the correct partition, which bounds how far a reader can be misled. Complexity trivial on judge-r2's own finding that the fix is a single clause at report lines 644-648. Likelihood held at medium-high: the stale clause is unchanged and blue took no round-3 turn.)
required_fix: Either (a) partition the sweep by session type and restrict the causal claim to the non-interactive share, stating what explains the interactive share (including "unresolved", which re-opens the serialization hypothesis there); or (b) show the interactive-looking transcripts were non-interactive and say how that was determined. Catechism 3(b) and §2 must stop asserting a single guard as the cause of all 5,754 blocks. CLASS RULE: every causal claim must have its trigger condition checked against the full population it is asserted over. Enumeration open.
acceptance_check: DOCUMENT-PROBE: read Catechism 3(b), §2 and the Provenance adjudication. The causal claim must name the session-type population it covers; if it covers the interactive share a mechanism must be stated; and the parsimony argument against serialization must be re-derived for that share or withdrawn for it. No LIVE-PROBE demanded: partitioning the existing store is a read.

### R2-3 — Verified at the leaf by red-merge-r2: WebFetch of arxiv.org/html/2603.05488v4 Table 1 returns DeepSe
§2 "Even when captured, the trace is not the reasoning" — "Independent probing work on a reasoning model reports performativity rates of 0.417 on MMLU (a recall task) and 0.012 on GPQA-Diamond (a multihop reasoning task) ... Performativity collapses across task difficulty."
severity medium-high | medium-high x medium | cx trivial | class within-source-condition-misattribution | supersedes R1-1 | found_by L1-F1
required_fix: Attribute both figures to DeepSeek-R1 671B by name at the prose site and in the footnote; report the GPT-OSS 120B row (0.435/0.227); restate the generalization to what two arms support, or drop "collapses". CLASS RULE: every quoted experimental result must be pinned to the model/benchmark/condition arm that produced it, and any generalization over a multi-arm source must be checked against every arm. Enumeration open.
acceptance_check: DOCUMENT-PROBE: read §2 and [^ReasoningTheater]; both model rows must appear with their own numbers, 0.417/0.012 attributed to DeepSeek-R1 by name, and no sentence may assert a collapse across task difficulty without reconciling GPT-OSS 120B 0.435 -> 0.227.

### R2-4 — R1-4 repair retired the nonexistent dev.to article and deleted the [^ToolTruncation] definition (blu
Catechism answer 4, final bullet — "Risk-accepted residuals we are choosing to live with, named: silent tool-result truncation with no audit marker;[^ToolTruncation]"
severity low | medium x low | cx trivial | class incomplete-repair-lag | supersedes R1-4 | found_by L5-F1
required_fix: Point the marker at [^ToolTruncationLimits] or remove it. CLASS RULE: every footnote retirement is followed by a report-wide grep of the retired label in BOTH directions — no orphan references, no orphan definitions. Enumeration open: the same sweep applies to [^NISTInitiative], retired in the same edit.
acceptance_check: DOCUMENT-PROBE: grep -n "ToolTruncation]" blue/report.md; every reference marker must have a matching definition and no definition may be unreferenced. Run the same for NISTInitiative.

### R2-5 — (i) R1-5 retired meta-intelligence.tech because the URL carried none of the content attributed to it
Footnote [^NISTAuditRequirement] — "AI Agent Governance and Compliance in 2026 — zylos.ai/research/2026-05-01-ai-agent-governance-compliance-2026/. Reports NIST-aligned guidance calling for structured audit logs to every agent action, logging the full chain: input, reasoning steps, tool calls, data accessed, output, and human approval." And open question 7 — "If NIST Q4 2026 interoperability profile mandates reasoning-step logging..."
severity medium | medium x medium | cx low | class repair-regression-citation | supersedes R1-5 | found_by L5-F2
required_fix: Re-source to a nist.gov primary carrying a reasoning-step logging requirement, or drop the NIST attribution and state what the secondary source supports — an emerging INDUSTRY audit-record schema with a reasoning-trace field that no Claude Code surface provides. Then strike or source "Q4 2026" at open question 7. CLASS RULE: a citation retired for not carrying its claim must have its replacement leaf-fetched and the quoted string confirmed present BEFORE the repair is announced; a dated-specific retirement is verified by grepping the date report-wide. Enumeration open.
acceptance_check: DOCUMENT-PROBE: fetch the URL at [^NISTAuditRequirement] and confirm the quoted sentence appears verbatim and is attributed there to NIST; if not, confirm footnote and §8 no longer attribute it to NIST. Then grep -n "Q4 2026" blue/report.md — any hit must carry a source stating it.

### R2-6 — R1-6 closed by replacing the contradicted 260+ with ~30 and disclosing the conflict. The disclosure 
Footnote [^ComplianceAPI] — "Access the Compliance API — support.claude.com article 13015708, and Compliance API — platform.claude.com. Publicly documented surface shows roughly 30 activity types including Claude Code events." And the §3 table cell — "~30 documented activity types, none reasoning (no 260+ category exposed in public documentation)".
severity low-medium | medium x low-medium | cx trivial | class repair-regression-citation | supersedes R1-6 | found_by L1-F1
required_fix: Cite the ~30 figure to the source that carries it (generalanalysis.com, labelled secondary) or drop the count from the cell and keep only the negative; repair the parenthetical to name the conflict as a COUNT conflict. CLASS RULE: when a figure is replaced during a repair the replacement carries its own citation duty — the footnote source list must be re-checked against the new number, never inherited from the old. Enumeration open.
acceptance_check: DOCUMENT-PROBE: read [^ComplianceAPI]; the ~30 figure must name a source that carries it, or be absent from the table cell. Fetch each URL the footnote names and confirm it supports what the footnote says it supports.

### R3-1 — LINEAGE DECLARED: this gap amends two closures judge-r2 entered in round 2 (R2-1, R2-5). Both archiv
Provenance and limitations of this round — "Not verified this round: … the arXiv identifiers in [^AgentBenches], **the NIST quotation's primary source**, and the vendor sales channel for raw thinking."
severity medium-high | medium-high x medium | cx trivial | class incomplete-repair-lag | supersedes R2-1 R2-5 | found_by L5-F2
required_fix: Reconcile the Provenance section against the current body and add the section to the standing propagation checklist. Specifically: strike 'the NIST quotation's primary source' from the not-verified list (or restore the claim it disclaims); and state the sweep-independence question at the same confidence in Provenance and §2 — resolved, or open with the outstanding confirmation named, at both sites. CLASS RULE: the Provenance/limitations section is a propagation site like any other; every future repair's site sweep must include it, and any hedge, retraction or not-verified entry parked there must be re-checked against the body whenever the claim it covers is edited. The enumeration of stale Provenance statements is declared OPEN — these are the ones red found, not a closed list, and the fix is the sweep, not the two edits.
acceptance_check: DOCUMENT-PROBE. (a) grep -n 'NIST' blue/report.md: no hit in the not-verified list may name a NIST quotation the report does not contain, and no hit may state unverified what [^NISTAuditRequirement] states verified. (b) grep -n 'pending confirmation|provisionally' blue/report.md: no hit may describe the 287/5,569 question as open while §2 describes it as answered. (c) blue's round-3 CHANGELOG propagation list must name the Provenance section — this is how red checks the CLASS rule landed rather than the two instances.

### R3-2 — This sentence is new in the R2-2 repair, and both the bench and red credited that repair without che
§2 "The mechanism, read out of the shipped client" — "The report's own §1 counts 16 top-level transcripts (interactive parent sessions) out of 294 files, meaning 278 are deeper-nested subagent and workflow runs."
severity medium | medium x medium | cx low | class numeric-collision-under-partition
required_fix: Either (a) partition the block count by the property that decides the branch — re-run the sweep grouped by session type, or by the depth proxy with the proxy's soundness argued — and state how many of the 5,754 blocks sit on each share; or (b) state plainly that the split is unquantified, that the 278-with-thinking and 278-nested figures are distinct counts that coincide, and that the interactive share's block count is unmeasured. In either case stop attributing the 'interactive parent sessions' gloss to §1, which does not make it. CLASS RULE: every figure reused across sections must be checked for set identity, not numeral identity, before an inference is drawn from the match; and any proxy standing in for the property a claim turns on (here nesting depth for session type) must be named as a proxy with its soundness argued at the site of use. Enumeration open.
acceptance_check: DOCUMENT-PROBE — read §2 lines 184-215. The two 278s must be distinguished in the text (or one replaced by a measured figure); the interactive share's block count must be given or explicitly declared unmeasured; and the 'interactive parent sessions' gloss must be dropped, sourced to a stated depth-to-session-type argument, or attributed to something other than §1. A LIVE-PROBE (re-running the store sweep grouped by session type) would discharge option (a) but is NOT demanded: option (b) is a document edit and is a complete answer.

## CLOSURE INDEX

R1-1 | closed_with_regression | placeholder | R2-3
R1-2 | closed_with_regression | Verified at the leaf: both pinned files EXIST in git at the  | R2-1
R1-3 | closed | False at the leaf. `feov-record blue --help` (plugin 0.10.0, | -
R1-4 | closed_with_regression | The author exists and is reachable: dev.to/api/articles?user | R2-4
R1-5 | closed_with_regression | Verified by direct fetch 2026-07-19: meta-intelligence.tech  | R2-5
R1-6 | closed_with_regression | The only accessible cited source (generalanalysis.com/guides | R2-6
R1-7 | closed_with_regression | The headline is a universal over Claude Code; its ground is  | R2-2
R1-8 | closed | The absolute ever contradicts the report own version-binding | -
R1-10 | closed | The four tiers grade atomic observations, but the report giv | -

## render anomalies (never silently normalized)

- duplicate key dedup'd: blue-lane-1:revision (nonce d3450f02)
- duplicate key dedup'd: red-lens-r1-L1:cite:platform.claude.com/docs/extended-thinking 2026-07-18 (nonce 805e7ff4)
- duplicate key dedup'd: red-lens-r1-L1:cite:code.claude.com/docs/en/agent-sdk/observability 2026-07-18 (nonce 805e7ff4)


## undisposed lens observations (every observation demands a merge disposition)

- red-lens-r1-L1 red-lens-r1-L1:observe:#1: Verified sections 1-5 (Transcript, Reasoning Channel, Settings, OpenTelemetry, Faithfulness): ~36 claims. All vendor-doc
- red-lens-r1-L1 red-lens-r1-L1:finding:#1: 
- red-lens-r1-L1 red-lens-r1-L1:finding:#2: 
- red-lens-r1-L2 L2-audit-complete: Sections 6-10 citation audit complete. 6 arXiv papers HIGH verified. 3 low-confidence findings recorded (unreachable sou
- red-lens-r2-L1 red-lens-r2-L1:observe:#1: Spot-checking local sweep. Store grew 294→306 files, 5754→6057 blocks (2026-07-18→2026-07-19). Permission gate on ~/.cla
- red-lens-r2-L1 red-lens-r2-L1:observe:#2: Compliance API docs (support.claude.com/article/13015708, platform.claude.com/docs/reference/compliance-api) now return 
- red-lens-r2-L5 red-lens-r2-L5:observe:#1: beginning round 2 logic/completeness audit of round 1 repairs
- red-lens-r2-L5 L5-F1: Broken footnote reference: [^ToolTruncation] cited at line 92 in 'Risk-accepted residuals' list but footnote definition 
- red-lens-r2-L5 L5-F2: Incomplete R1-5 repair: Open question 7 references 'NIST's Q4 2026 interoperability profile' without sourcing. R1-5 requ
- red-lens-r2-L5 L5-F3: Logic gap in soundness tier definition: §6 Tier 1 lists 'permission approvals and denials' as directly observable facts,
- red-lens-r2-L5 red-lens-r2-L5:observe:#2: audit complete: verified all 10 R1 repairs implemented; found 3 logic/consistency issues: broken footnote reference, inc
- red-lens-r3-L1 red-lens-r3-L1:observe:#1: Round 3 audit of slice 1 (Catechism + §1-5) complete. 35 claims verified; 4 NEW citations corroborated; spot-checks on p
- red-lens-r3-L2 audit-scope: Slice 2 (L2-F): §6-§10, open questions, provenance. Focus: changed sections per CHANGELOG (§6 composition rule R1-10; §8
- red-lens-r3-L5 L5-F1: The test to set showThinkingSummaries=true on a non-interactive session is the single strongest counterexample to the he
- red-lens-r3-L5 L5-F2: The report acknowledges but does not resolve a competing hypothesis: serialization (API returns empty content, client dr
- red-lens-r3-L5 L5-F3: The report uses issue #32810 (closed not-planned on v2.1.71) to establish the v2.1.71 mechanism, then shows the mechanis
- red-lens-r3-L5 L5-F4: The risk matrix prescribes LOW complexity for tier-label enforcement. But the report marks several figures as 'not leaf-
- red-lens-r3-L5 L5-F5: The faithfulness case rests on: (1) vendor self-report (Anthropic Alignment Science, generic), (2) performativity resear
- red-lens-r3-L5 L5-F6: The report's case for artifacts assumes recording is DURING work, not post-hoc. But the report does not discuss timing o
- red-lens-r3-L5 L5-F7: The report conflates 'effort level is not exposed' with 'reasoning cannot be audited.' But the recommendation is to audi
- red-lens-r3-L5 L5-F8: Properly scoped and honestly caveated. BUT: the Compliance API discrepancy is unresolved — lane-1 reported 260+ activity
- red-lens-r3-L5 L5-F9: Enforcing tier discipline is not LOW complexity if the report itself conflates tiers in its own recommendations. The art
- red-lens-r3-L6 L6-F1: Report argues thinking blocks are unfaithful (post-hoc), then recommends artifact recording as primary channel—equally s
- red-lens-r3-L6 L6-F2: Report compares recording costs (artifact=one line, thinking=indefinite maintenance). But §6 places 'agent reasoned that
- red-lens-r3-L6 L6-F3: Report accepts truncation risk 'with disclosure' but proposes no post-hoc detection mechanism. Risk matrix says 'no audi
- red-lens-r3-L6 L6-F4: Report identifies OQ1 as critical-path experiment: does showThinkingSummaries:true override the force-omit guard on non-
- red-lens-r3-L6 L6-F5: Report is version-bound to v2.1.215 read 2026-07-19. Notes history of vendor changes via server-side flag (tengu_quiet_h
- red-lens-r3-L6 L6-F6: Report concludes 'no reasoning category' based on public surface (30 types enumerated). Lane-1 reported 260+ types, unve
