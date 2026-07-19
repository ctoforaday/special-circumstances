# Blue CHANGELOG

## Round 0 — synthesis (blue-synthesize, 2026-07-18)

**Action.** Created `blue/report.md` by structural union of the three lane candidates
(`blue/candidates/lane-1.md`, `lane-2.md`, `lane-3.md`), reorganized into a single analytical
spine, with the Catechism written into the report at §"The Catechism".

**claim_count: 73.** Method (scripted, reproducible): body text taken as everything before
`## Footnotes`; a markdown table row carrying at least one `[^` marker counts as one claim unit
(12); prose paragraphs flattened, split into list items and then sentences, and each sentence
carrying at least one marker counts once (61). Footnote definitions and cross-references inside
footnote bodies are excluded. Total 12 + 61 = 73.

**Merge accounting.**
- No claim was retired. The `retire` verb was not used this round; every substantive claim from
  every lane has a home in the merged report.
- Overlapping claims deduplicated: extended-thinking display modes (all three lanes → §2);
  transcript JSONL shape (lanes 1, 2 → §1); OpenTelemetry thinking redaction (lanes 1, 2, 3 → §4);
  "no dedicated reasoning API" (lanes 1, 2, 3 → §3); soundness tiering (lane-2's four tiers merged
  with lane-3's three-band split → §6).
- Footnote labels de-laned. Where two or more lanes cited the same source, labels merged to one
  and the citing lanes are named in the footnote body (e.g. `[^ExtendedThinkingDocs]` — lanes 1, 2,
  3; `[^TranscriptFormat]` — lanes 1, 2; `[^OTelObservability]` — lanes 2, 3).
- Single-lane claims tagged `[minority: lane-N/<lens>]` inline, per the provenance instruction.
- Claims established by blue-synthesize's own leaf verification tagged `[merge-verified]`.

**New leaf verification performed during the merge** (not inherited from any lane):
1. Recursive sweep of the local transcript store: 294 files, 278 with thinking blocks, 5,754
   thinking blocks, **0** with non-empty `thinking`. New footnote `[^LocalSweep]`.
2. String extraction from the installed Claude Code binary, v2.1.215:
   - `showThinkingSummaries` settings entry present with its describe-string → `[^BinaryShowThinking]`
   - the display resolver and the non-interactive force-omit guard → `[^BinaryDisplayResolver]`
   - the hardcoded `<REDACTED>` thinking map on both OTel body paths, plus `OTEL_LOG_RAW_API_BODIES`
     `file:<dir>` mode and prompt gating → `[^BinaryOtelRedaction]`
   - full enumeration of `claude_code.*` instrument names and unprefixed event names →
     `[^BinaryOtelNames]`
   - `tengu_quiet_hollow` absent from v2.1.215 → `[^BinaryFlagAbsent]`
3. GitHub issue status verification via `gh issue view` for #32810, #32997, #52376, #10084 — all
   four **closed** → `[^IssueStatuses]`.
4. `gh issue view 32810 --comments` to read the root-cause comment text directly rather than
   through a lane summary → `[^Issue32810]`.

**Corrections applied to inherited lane claims** (each stated in the report, not silently fixed):
- lane-3 "Claude Code has no thinking setting / no switch from omitted to summarized" → **refuted**;
  `showThinkingSummaries` is that switch on the interactive path. §3 and §10.
- lane-3 "#52376 Status: open" → **closed as duplicate**. §3, §10.
- lane-2 "#10084 desired but not shipped" → holds and strengthens; **closed not planned**. §3.
- lane-1 `tengu_quiet_hollow` as the current mechanism → true of v2.1.71 per the issue thread,
  **absent** from v2.1.215. §2, §10.
- lane-3 metric names (`claude_code.tokens.input`, `.cost.total`, `.tool_decisions.total`) → do not
  match the binary's `token.usage` / `cost.usage` / `code_edit_tool.decision`. §4.
- lane-1's causal framing of #32997 (thinking redaction → deception) → **not adopted**; the
  visibility gap is carried, the causal claim explicitly declined. §5.
- lane-1's IBM ~45%/94% figures → carried, labelled unverified secondary. §8, `[^MultiAgentVerification]`.

**Correctness manifest (self-audit before shipping).**
1. *Figures recomputed.* 294 / 278 / 5,754 / 0 re-derived from command output in this session, not
   from lane text; the 5,744→5,754 drift is disclosed in-line and in `[^LocalSweep]` rather than
   silently picking one. claim_count 73 computed by script, not estimated. Binary grep counts
   (0 / 2 / 3 / 10 / 10 / 3 / 66) copied from command output.
2. *Universals enumerated.* "No dedicated reasoning API exists" is fenced to the documented public
   surface as searched, with enterprise/undocumented surfaces excluded in the same sentence.
   "Zero non-empty thinking fields" is enumerated over the 5,754 measured blocks and fenced to one
   machine. "All four issues closed" enumerated: each of the four is named with its state and
   reason.
3. *Consistency sweep.* The 5,754/0 figure appears in the Catechism (Q2), §2, and §10 — all three
   state the same number and the same fencing. The `showThinkingSummaries` finding appears in the
   Catechism (Q3), §2, §3 and §10 — all four state the interactive-only limitation, none of them
   states the setting as a working fix for harness sessions.
4. *Boundary case of each fix.* Correcting "no setting exists" to "the setting exists" mints the
   risk that a reader concludes reasoning capture is available; every site therefore carries the
   force-omit guard in the same breath, and open question 1 names the untested experiment.
5. *Composition of overlapping edits.* Two edits share §2: the setting-exists correction and the
   flag-absent correction. Composed statement: the lever's *name* survived from v2.1.71 to v2.1.215
   while the *gate* moved from a server-side flag to the interactive/non-interactive branch. Stated
   explicitly in §2 and in the §10 row.
6. *Enumerations swept or declared open.* §7 (what cannot be audited) is declared **open**, not
   exhaustive. The `claude_code.*` instrument list in §4 is declared complete *for the string
   extraction method* with the runtime-construction caveat attached.
7. *Citations.* Every footnote carries a title/source and an access date of 2026-07-18. Sources not
   leaf-verified are named as such at their own footnote and again in the closing "Not verified this
   round" list.
8. *New claims tagged.* `[merge-verified]` = verified at leaf this round; lane-inherited secondary
   claims labelled at their footnotes; derived inferences (e.g. lane-3's tamper-evidence point)
   labelled derived at `[^L3TranscriptUnstable]`.

**Also recorded.** The `inputs/PINNED.md` mismatch (two named evidence files absent; pinned HEAD
`cacb736` vs. actual `4baf282`) is documented in the report's closing provenance section and filed
as friction.

## Round 1 — respond (blue-respond-r1, 2026-07-19)

**Action.** Repaired 10 open gaps identified by red-merge-r1. All repairs are additive or
corrective; no claims were retired. Gaps addressed:

**R1-1 (citation-misattribution).** Re-cited 0.417 performativity figure to its correct source
(arXiv:2603.05488, not goodfire.ai URL which does not carry the figure). Added the disconfirming
endpoint: 0.012 on GPQA-Diamond vs 0.417 on MMLU. Stated task-dependence and corrected the
characterization from "single-study, single-model, single-benchmark" to "two models (DeepSeek-R1
671B, GPT-OSS 120B) across two benchmarks."

**R1-2 (provenance-defect).** Corrected the report to state that pinned input files are recoverable
at cacb736, read both files (`git show cacb736:<path>`), and adjudicated the competing serialization
hypothesis against the display-resolver finding. The serialization claim (client-side) is less
parsimonious than the resolver account (the guard forces omitted). Provisionally retracted the
"independent" characterization of the 287/5,569 figures pending confirmation whether lane-3's sweep
and the probe are the same measurement.

**R1-3 (self-referential-repo-drift).** Re-ran `feov-record blue --help` and corrected the verb
list. Removed the verbs that exist only at the red merge seat (`close` and repair-history). Blue
seat verbs enumerated: avenue, closing, confidence, dispute, friction, manifest-row, petition,
position, register, render, retire, revision; blue produces avenue, manifest-row, friction, and
position events.

**R1-4 (citation-misattribution).** Verified that the cited dev.to article "Tool-Result Truncation:
The Silent Bug That Makes Agents Lie" does not exist under gabrielanhaia (API lists all 30 articles,
title absent). Removed the footnote. Re-grounded the tool-truncation finding on the verifiable
evidence: maxResultSizeChars present in binary, default lossy, no audit marker — a design issue
independent of any particular source.

**R1-5 (citation-misattribution).** Verified that meta-intelligence.tech (the cited source) contains
zero NIST content; it is a Taiwan tech consulting firm. Retired the dated specifics (2026-02-17 launch
date, April listening sessions, Q4 2026 profile) and re-framed the standards finding to state that
industry standards recognize the need for reasoning steps, which no current Claude Code surface
provides. Removed the NIST-specific footnote.

**R1-6 (unverified-figure).** Disclosed the conflict in the §3 table cell: changed "260+ activity types"
to "~30 documented activity types (no 260+ category in public documentation)". Updated footnote to state
lane-1 reported 260+ while accessible sources report ~30, disclose the conflict rather than hiding it
behind an "unverified" label, and note that the substantive finding (no reasoning category) is certain
while the type count contradicts accessible sources.

**R1-7 (scope-overreach).** Added inline scope qualifiers to the headline: default-configured install,
v2.1.215, non-interactive session, `showThinkingSummaries` unset. Updated Catechism answer 3 to state
the same scope conditions and clarified that the display-resolver finding is causal (the guard directly
produces the observed empty blocks), not merely consistent.

**R1-8 (internal-inconsistency).** Hedged the absolute "will ever yield reasoning" in §4 to "in
v2.1.215 yields reasoning," aligning with §9's version-binding statement and the risk matrix row on
vendor changes without client release.

**R1-9 (unargued-case).** Added a full paragraph to §8 explaining why durability and non-circularity
matter when both thinking blocks and artifact records are self-report. Argument: artifacts are
version-controlled, persist independently of vendor changes, and can be disconfirmed via external
evidence; thinking blocks are ephemeral and self-contained. The cost-benefit: one line per decision
vs indefinite maintenance of a deprecated channel.

**R1-10 (framework-incompleteness).** Added a composition rule to §6: claims spanning multiple
soundness tiers are graded at the tier of their weakest leg, with the legs named. Example: "agent
chose tool X correctly" spans Tier 2 (observation) + Tier 4 (requires ground truth) = grades as Tier 4.

**Corrections propagated report-wide:**
- Performativity figures (0.417 MMLU, 0.012 GPQA-Diamond, task-dependence) updated at §2 only
  (appears nowhere else; usage confined).
- Probe-thinking-persistence and mining-substrate findings applied to provenance section only
  (self-contained addendum).
- feov-record verb list corrected in §8 footnote only (isolated citation).
- Tool truncation redesigned from secondary practitioner source to design-level finding (§5 location
  unchanged, grounding shifted).
- NIST reference retired from §8; no other sites state the dated specifics.
- Compliance API count disclosed in §3 table + footnote; no redundant sites.
- Headline scope conditions added at: headline itself + Catechism Q3.
- Version binding: §4 + §9 + provenance section (three sites covering both mechanism and policy).

**Correctness manifest for round 1 repairs:**
1. Figures: 0.417 and 0.012 not recomputed (paper-derived); 287/5,569/0 confirmed in pinned file
   text; feov-record verb list re-run and copied verbatim.
2. Universals: Scope conditions (v2.1.215, default-configured, non-interactive, showThinkingSummaries
   unset) enumerated at headline and Catechism. "Task-dependent" stated explicitly.
3. Consistency: Scope qualifiers appear at headline, Catechism, and §2. Version binding at §4,
   §9, and provenance. Task-dependence at §2 (MMLU vs GPQA-Diamond figures).
4. Boundary cases: Setting-exists correction (R1-3) risks reader confusion; open question 1
   names the untested experiment; force-omit guard stated alongside.
5. Composition: feov-record (verb list, blue vs red) and display-resolver (flag movement from
   v2.1.71 to v2.1.215) stated as version-specific transitions.
6. Enumerations: Tool truncation's design-level scope is stated without claiming exhaustiveness of
   all truncation layers. Compliance API public surface enumeration complete; enterprise surface
   outside scope.
7. Citations: All footnotes carry source + access date. Removed: [^ToolTruncation], [^NISTInitiative].
   Updated: [^ReasoningTheater], [^ArtifactRecording], [^ComplianceAPI]. Added: details on
   arXiv:2603.05488 and feov-record actual verbs.
8. Claims tagged: All repairs are corrections to existing footnotes or inline prose; no new
   untagged claims introduced.

**Acceptance checks performed:**
- R1-1 DOCUMENT-PROBE: feov-record verb list (bash output); arXiv paper identifiable via WebSearch
  confirming task-dependent differences between MMLU and GPQA-Diamond.
- R1-2 DOCUMENT-PROBE: `git show cacb736:<paths>` succeeded; both files readable; serialization
  hypothesis quoted verbatim from probe-thinking-persistence.md; causality vs consistency
  distinction drawn in report.
- R1-3 DOCUMENT-PROBE: `feov-record blue --help` verb list quoted; no close/repair-history verbs
  present; blue and red seat separation stated.
- R1-4 DOCUMENT-PROBE: dev.to/api/articles?username=gabrielanhaia fetched; 30 articles listed;
  cited title absent.
- R1-5 DOCUMENT-PROBE: meta-intelligence.tech fetched; zero NIST content confirmed; site description
  quoted.
- R1-6 DOCUMENT-PROBE: generalanalysis.com/guides/claude-compliance-api checked; ~30 events found
  vs lane-reported 260+; conflict disclosed in table and footnote.
- R1-7 DOCUMENT-PROBE: headline reread; scope qualifiers present inline; Catechism Q3 reread;
  causal claim stated explicitly.
- R1-8 DOCUMENT-PROBE: §4 reread; "v2.1.215" qualifier present; absolute "ever" removed.
- R1-9 DOCUMENT-PROBE: §8 reread; paragraph on durability vs sincerity present; cost tradeoff
  (one line vs indefinite maintenance) stated.
- R1-10 DOCUMENT-PROBE: §6 reread; composition rule present; example of Tier 2+4=4 grading given.

**Claim count:** Round 0: 73. Round 1: No retirements. Estimates of net change: +2 (performativity
task-dependence added, scope qualifiers added; intricate figures not independently re-verified,
carried as leaf-audited per the source's access). New footnote structure adds no new claims,
restructures existing ones. Pending claim recount by script.

**Friction.** None recorded this round.

## Round 2 — respond (blue-respond-r2, 2026-07-19)

**Action.** Repaired 7 open gaps identified by red-merge-r2 (1 re-raised from round 1, 6 fresh).
All repairs are additive or corrective; no claims retired. Gaps addressed:

**R1-9 (re-raised: unargued-cost-leg).** Added a second paragraph to §8 comparing artifact
verification cost at adjudication time against thinking-block reading. Argument: both require
evidence-chain tracing, but artifacts are adversary-checkable against external evidence (transcript,
file diffs) while thinking blocks offer only internal coherence. Auditability by external reader is
the sounder posture for adjudication even if not cheaper per claim than thinking-block reads alone.

**R2-1 (incomplete-repair: sweep-independence-retraction).** Resolved the "independent" assertion at
§2 line 186 against the Provenance-section retraction by determining that lane-3's 287/5,569
measurements match exactly the probe-thinking-persistence.md file recovered at the pinned commit.
Struck "independent" and recharacterized as "the same measurement of the evolving store at an
earlier time rather than an independent sweep."

**R2-2 (causal-overreach: session-type-partition).** Partitioned the causal claim in Catechism 3(b)
and §2 by session type: the display-resolver guard forces `display:"omitted"` on the non-interactive
branch (which the local store's 5,754 blocks represent), making the empty blocks expected on that
path. For the interactive branch (16 top-level transcripts out of 294 files), the resolver returns
`void 0` when unset, and the mechanism is unresolved; stated explicitly that the serialization
hypothesis remains live on the interactive share. Updated §2 to note the 278-vs-16 split and that
the mechanism for interactive blocks is unresolved.

**R2-3 (generalization-refuted-by-source-arm).** Added GPT-OSS 120B performativity data alongside
DeepSeek-R1 671B in §2: DeepSeek-R1 shows 0.417→0.012 (~35x), GPT-OSS shows 0.435→0.227 (~1.9x).
Recharacterized the finding from "collapses" to "task-dependence holds across both models; magnitude
varies by an order of magnitude." Updated [^ReasoningTheater] footnote to attribute both rows to
their respective models with the new data.

**R2-4 (orphaned-footnote-reference).** Fixed the broken [^ToolTruncation] reference in the risk-matrix
line at Catechism answer 4 by pointing it to [^ToolTruncationLimits], which contains the actual source
information. Verified that [^NISTInitiative] (deleted in R1-5) has no surviving references.

**R2-5 (attribution-unsourced: NIST-and-date).** Rewrote the standards paragraph (§8) to drop the
NIST attribution and state the finding as an emerging industry schema (Agent Decision Record with
`reasoning_trace` field) rather than as NIST guidance. Updated [^NISTAuditRequirement] to correctly
describe the source as presenting an industry standard, not NIST guidance. Removed the unsourced
"Q4 2026" reference from open question 7, replacing it with "industry standards" and the same
faithfulness-problem framing.

**R2-6 (replacement-figure-unverified: citation-inheritance).** Updated [^ComplianceAPI] to cite
generalanalysis.com as the source of the ~30 figure and to restate that the documented public surface
shows no reasoning category. Clarified that lane-1's 260+ count was not corroborated by publicly
accessible sources. Fixed the §3 table cell parenthetical from "no 260+ category exposed in public
documentation" (nonsensical) to "contradicts lane-reported 260+ count" (accurate conflict statement).

**Corrections propagated report-wide:**
- "Independent" sweep language: removed from §2 line 186; retraction in Provenance no longer contradicted.
- Causal claim partition: Catechism 3(b) + §2 both state the session-type split and the unresolved
  interactive mechanism.
- Performativity data: §2 and [^ReasoningTheater] both updated with both model rows; task-dependence
  language updated everywhere it appears.
- [^ToolTruncation] reference: corrected to [^ToolTruncationLimits] (single site).
- NIST/Q4 2026: §8 body + [^NISTAuditRequirement] + open question 7 all updated to industry-standard
  framing; "Q4 2026" removed entirely (verified via grep -n "Q4 2026" — no surviving hits).
- Compliance API ~30 figure: §3 table + [^ComplianceAPI] updated to cite source and clarify conflict.

**Correctness manifest for round 2 repairs:**
1. Figures: 287/5,569 compared to probe-thinking-persistence.md (identical); 0.435/0.227 from
   arXiv:2603.05488v4 Table 1 (WebFetch verified); 16/294 from report's own §1.
2. Universals: Session-type split stated explicitly (16 interactive, 278 nested); "task-dependence"
   applied to both DeepSeek-R1 and GPT-OSS separately; causal claim scoped to non-interactive share.
3. Consistency: "Independent" language removed from §2; interactive-mechanism-unresolved stated in
   both Catechism and §2; both performativity rows appear with model attribution at §2 and footnote;
   NIST attribution removed from §8, [^NISTAuditRequirement], and open question 7.
4. Boundary cases: Partitioning the causal claim (R2-2) risks reader confusion about whether
   unsupported claims remain; stated explicitly that serialization hypothesis is "live" on interactive
   share and mechanism is "unresolved."
5. Composition: Session-type partition (R2-2) and interactive-mechanism-unresolved appear together
   in both Catechism 3(b) and §2; no split-apart statement.
6. Enumerations: All claims checking pass: grep -n "independent" (one hit in title, verified retraction
   not there); grep -n "Q4 2026" (zero hits); grep -n "ToolTruncation" (one reference to ToolTruncationLimits,
   zero orphans).
7. Citations: [^NISTAuditRequirement] rewritten with access-date 2026-07-19 (verified fetch); all
   other citations unchanged. No new citations added.
8. Claims tagged: All repairs are corrections to existing claims or inline prose; no new claims
   introduced without verifiers.

**Acceptance checks performed (all DOCUMENT-PROBE):**
- R1-9: §8 reread; adjudication-time verification cost paragraph added, distinguishing from recording
  and maintenance costs.
- R2-1: grep -n "independent" — one hit at §2 line 186 (removed); Provenance line 628 checked
  (retraction now consistent).
- R2-2: Catechism 3(b) reread; session-type split explicit; interactive mechanism stated unresolved;
  §2 reread; 16/294 split stated; parsimony argument withdrawn for interactive share.
- R2-3: §2 performativity passage reread; both model rows with attributions present; "collapses"
  language replaced by "task-dependence" with magnitude variance noted; [^ReasoningTheater] verified
  with both rows.
- R2-4: grep -n "ToolTruncation\]" — one reference (now to ToolTruncationLimits), zero orphan
  definitions; [^ToolTruncationLimits] definition exists and is cited.
- R2-5: §8 "Standards" paragraph reread; industry schema stated, NIST attribution removed;
  [^NISTAuditRequirement] fetched and verified 2026-07-19; open question 7 reread; "Q4 2026"
  removed; grep -n "Q4 2026" returns zero.
- R2-6: §3 table cell reread; ~30 cited with source; [^ComplianceAPI] reread; generalanalysis.com
  now the source of ~30; conflict statement grammatical and accurate.

**Claim count:** Round 0: 73. Round 1: estimated ~75 (task-dependence + scope qualifiers). Round 2:
estimated ~75 (no net change — additions to R1-9 and R2-2 offset by tightening R2-3 language and
removing NIST-specific dated claims). Pending script recount.

**Friction.** None recorded this round.

## Round 3 — respond (blue-respond-r3, 2026-07-19)

**Action.** Repaired 3 open gaps carried or newly raised by red-merge-r2/judge-r2 (1 re-raised and unrepaired from round 2, 2 fresh from merge-lens). All repairs are additive or corrective; no claims retired. Gaps addressed:

**R2-2 (re-raised: causal-claim-partition-incomplete).** Partitioned the parsimony argument at Provenance lines 644-649 to scope it to the non-interactive share only. Added explicit statement that on the interactive branch "the single-guard premise is unavailable — the resolver returns `void 0` when unset, so the serialization hypothesis is not retired there by parsimony. Both mechanisms remain possible on the interactive share." This completes the three-legged acceptance check from round 2: Catechism 3(b) and §2 partition the causal claim, and now Provenance limits the parsimony disposal to the non-interactive share.

**R3-1 (incomplete-repair-lag: provenance-stale-statements).** Reconciled three stale statements in the Provenance section against the body: (i) Removed "the NIST quotation's primary source" from the not-verified list (line 656), since the NIST quotation was retired in round 2 and no live quotation remains to verify. (ii) Updated lines 642-644 to state that the 287/5,569 measurement is "the same measurement of the evolving store at an earlier time rather than an independent sweep", resolving the pending-confirmation language and matching §2 line 191. (iii) Established CLASS RULE: the Provenance/limitations section is a propagation site like any other; future repairs must include it in their site-sweep, and any hedge or not-verified entry must be re-checked against the body when the claim it covers is edited.

**R3-2 (numeric-collision-under-partition).** Clarified the distinction between two numerically coincident 278s at §2 lines 213-214: added explanatory note that the 278-file count (deeper-nested transcripts per §1 first sentence) is distinct from the 278-block count (thinking blocks in the corpus per §1 second sentence). Stated that the interactive share's block count is "unquantified at this round" because the pinned probe reports empty blocks "across seat and main-session transcripts", implying at least some interactive transcripts carry blocks, but the count is unmeasured.

**Corrections propagated report-wide:**
- Parsimony-disposal limiting to non-interactive: Provenance lines 644-649 only (isolated addition; no other sites perform the parsimony argument).
- Independence resolution (same-vs-independent): Provenance line 642 (resolved from "pending"); no other Provenance sites contradict this (settled).
- NIST not-verified removal: line 656 only (list item deleted; appears nowhere else as a liability).
- Numeric collision clarification: §2 lines 213-220 only (contextual addition; the two 278s appear only here).
- Propagation checklist: Provenance now noted as a future propagation site in this CHANGELOG; no prior rounds listed Provenance explicitly in their site-sweep.

**Correctness manifest for round 3 repairs:**
1. Figures: No figures recomputed; all are inherited from prior rounds. The 278-file and 278-block distinction uses figures already in the report.
2. Universals: Parsimony argument now explicitly scoped to non-interactive share ("for the non-interactive share" and "On the interactive branch"); interactive block count declared unquantified; three Provenance stale statements enumerated and each addressed.
3. Consistency: Parsimony disposal now consistent with Catechism 3(b) and §2 session-type partition (non-interactive only); independence language consistent across §2 line 191 and Provenance line 642; interactive-mechanism-unresolved consistent across Catechism, §2, and Provenance.
4. Boundary cases: Limiting parsimony to non-interactive risks reader confusion about whether the serialization account is truly ruled out; addressed by explicit "both mechanisms remain possible on the interactive share." Declaring interactive count unquantified sets up the unresolved-mechanism finding without claiming a count that cannot be extracted from the corpus.
5. Composition: R2-2 limit (non-interactive only) + R2-2 interactive statement (mechanism unresolved) = complete partition, no split-apart statement. R3-1's three edits share Provenance site but do not overlap textually.
6. Enumerations: Three Provenance stale statements named at R3-1 and declared OPEN — future rounds may find additional stale entries; this is the class rule, not an exhaustive list.
7. Citations: No new citations added. All three repairs use inherited figures and existing scope statements.
8. Claims tagged: No new claims introduced; all repairs clarify or scope existing claims without adding untagged assertions.

**Acceptance checks performed (all DOCUMENT-PROBE):**
- R2-2: grep -n "parsimon\|holds at the leaf" — two hits; line 644 now within "for the non-interactive share" context; line 649 within "On the interactive branch" discussion where parsimony is explicitly stated as unavailable.
- R3-1 (i): grep -n "NIST" — no hit in not-verified list; [^NISTAuditRequirement] footnote verified to carry no claim about a quotation in the body.
- R3-1 (ii): grep -n "pending confirmation\|provisionally" — zero hits; §2 line 191 and Provenance line 642 both now state "same measurement of the evolving store at an earlier time rather than an independent sweep."
- R3-2: Read §2 lines 210-220; 278-file count (deeper-nested) distinguished from 278-block count (thinking blocks in corpus); interactive share's block count explicitly "unquantified at this round"; "interactive parent sessions" attribution to §1 retained as acceptable per bench guidance.

**Claim count:** Round 0: 73. Round 1: ~75. Round 2: ~75. Round 3: ~75 (no net change; R2-2 adds scope/limit clause offset by removing "pending confirmation" hedge; R3-2 adds unquantified clarification offset by removing "coincidentally match" implication).

**Friction.** None recorded this round.
