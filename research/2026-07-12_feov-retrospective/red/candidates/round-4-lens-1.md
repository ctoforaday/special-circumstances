# Round 4 — Lens 1: leaf-node citation verification (slice: header + §0 + §1 "Doubts")

Scope: `blue/report.md` lines 1–444 (title block, §0 Headline, §1.1–1.4). Every footnote and
inline reference reachable from this slice was either (a) confirmed already at HIGH confidence in
`red/citation-ledger.md` with no CHANGELOG change to its section this round, and therefore not
re-fetched, or (b) independently re-verified below because it was not yet in the ledger or because
a live source moved. `blue/CHANGELOG.md` shows no Round 4 entry yet (last entry is Round 3) —
consistent with this being red's audit of the round-3 state.

## Verified this pass (new ledger entries — see citation-ledger.md)

- `[^LocalGrep]` zero-match grep claim — re-run independently: `research/2026-07-12_memory-
  architecture/blue/report.md` is exactly 2,145 lines; grep `lane-1|lane-2|lane 1|lane 2`
  (case-insensitive) returns 0 matches. **HIGH.**
- `[^BlueReportGrep]` "7 total matches... none per-claim" — re-run independently against both named
  files. `blue/report.md` alone: 5 matching lines (6, 7, 9, 721, 787). Assembled
  `research/2026-07-12_memory-architecture/report.md` alone: 7 matching lines (116, 117, 119, 831,
  897, 2417, 2963) — a superset containing all 5 of `blue/report.md`'s lines (renumbered) plus 2
  more (2417: red's own housekeeping note "Disconfirming budget met in both blue lanes. Not a gap";
  2963: `debate.md`'s Round-0 BLUE-opening method summary). Total *unique* matches = 7, exactly as
  claimed; both extra lines are method-level/housekeeping, not per-claim attribution. **HIGH**
  (initially looked like a miscount — 5+7=12 — until checked for the subset relationship; false
  alarm avoided by direct comparison of matched line content, not just counts).
- `[^LocalGrepRed]` "66 matches" for `corroborat` in `red/findings.md` — the Grep tool's `count`
  output mode returned 64 (it counts matching *lines*, not occurrences); re-ran with
  `grep -io "corroborat[a-z]*" | wc -l` (occurrence count) = 66, exact match. **HIGH** (tool-mode
  artifact, not a report defect — noting for future lens passes: use occurrence-count, not
  line-count, when a footnote's claimed number could plausibly be either).
- `[^RedFindingsGrep]` "1 match, line 156" — confirmed exact line and text match. **HIGH.**
- `[^ChangelogR0]` merge-vocabulary quotes — already HIGH in ledger (round 1); unchanged, frozen
  historical artifact (a different, completed research run's CHANGELOG). No re-fetch needed.
- `[^BlueReportUnverified]` §10 quote — confirmed verbatim at `blue/report.md` line 787, and
  confirmed that line falls structurally under `## 10. Unverified items (labeled, not laundered)`
  (section header at line 783). **HIGH.**
- `[^RedAuditorSpec]` — confirmed against `plugins/frank-exchange-of-views/agents/red-auditor.md`:
  line 6 `memory: project`, line 15 corroboration-confidence mandate, line 20 "AFTER catching a new
  gap *pattern* (not instance), YOU MUST record it in your project memory" — all three sub-claims
  verbatim. **HIGH.**
- `[^ClaimManifest]` — confirmed against `ideas/backlog.md` item (5): "CLAIM MANIFEST... One
  artifact, five wins" — verbatim (report's "one artifact, five wins" is a faithful lowercase
  paraphrase of the source's capitalized opening). **HIGH.**
- §1.4's "generalizing R4-1's denylist-vs-allowlist finding" characterization of
  `pattern_invariant_soundness_by_enumeration.md` — read the memory file directly: it does
  describe an R4-1 finding from "the memory-architecture debate" (a different project's round 4,
  not this retrospective's) about an outbound secret-gate wiring `Bash` while the inbound taint
  invariant omitted it, and does recommend "inverting to an allowlist." Report's summary is
  accurate. **HIGH.**
- §1.4's "unimplemented" characterization of `ideas/backlog.md` item 4 (blue pre-flight self-audit)
  — confirmed: item (4) exists, no `[x]`, still an open `- [ ]` item. **HIGH.**
- Gap-pattern-file count ("15... and eleven more") — **not re-raised.** Live count is now 23 files
  (24 entries in the memory directory minus `MEMORY.md`), matching red's own round-3 merge-time
  re-verification (23, up from 18 at round-2 audit time) already logged in the ledger and there
  explicitly dispositioned "non-gap drift." No further drift since round 3 (still 23). Confirming
  the round-3 disposition still holds rather than re-litigating a self-adjudicated non-gap.

## New finding

### R4-1 — HIGH — certain × high × low-medium: the live-cited backlog now diagnoses, by name, the exact structural bug behind this report's own three-round-recurring footnote defect, and blue's own repair never surfaces it

**Location:** §1.1, the `[^DiminishingReturns]` footnote's Round 3 (R3-9) text: *"This is the
second round this exact footnote required a citation-discipline fix and the third round running a
defect recurred in this one footnote (R1-5, R2-4, now R3-9)... a standing argument for the claim
manifest (§3 row 5) applying to blue's own footnotes, not only to cross-lane provenance."*

**What I found:** `ideas/backlog.md` — the same live file blue already cites in this report via
`[^CostFigureProvenance]` (pinned at `main` @ `d164ab2`, 2026-07-14 00:24:02 -0700) — gained a new
item at commit `42dba2d` (2026-07-14 00:49:14 -0700, **25 minutes after** blue's last pin, and
still current HEAD as of this audit): *"frank-exchange-of-views: the docket detector tracks IDs,
not lineages (discovered live in run 3, rounds 1-3): red closes gaps 'WITH REGRESSION' and mints
successor gaps under fresh IDs (R1-5 → R2-4 → R3-4/R3-9), so `prevGapIds.has(g.id)` never matches,
the contested docket never arms, and the judge never sees a dispute lineage no matter how long it
persists — the only remaining brake is the maxRounds cost ceiling."*

The gap-ID chain the backlog names — **R1-5 → R2-4 → R3-4/R3-9** — is not a hypothetical: it is
this retrospective's own red/blue history for this exact footnote, confirmed by direct read of
`blue/CHANGELOG.md` (R1-5 restates the "2–4 agents" plateau; R2-4 flags/rebuts the "7 agents"
clause; R3-4 fixes the §1.1 body-lags-footnote lag; R3-9 disambiguates R2-4's replacement sentence)
— all four gap IDs, all one footnote. I independently confirmed the mechanism named in the backlog
item by reading `debate.js` directly: line 178, `const contested = redEnv.gaps.filter(g =>
prevGapIds.has(g.id))` — contested-docket membership is keyed purely on gap `id` identity. When red
"deepens" a finding across rounds (legitimate re-verification, not spinning — the backlog item
itself says so), each round's gap object gets a fresh id (`R2-4` is not `R1-5`), so the filter never
matches, `contested` stays empty, the judge branch (`if (contested.length > 0)`) never fires for
this dispute, and the debate proceeds by silent blue-revise/red-re-flag cycling until `maxRounds` —
exactly what happened here, three times, to one footnote.

**Why this is a gap in the report, not just a fact about the world:** blue's own §1.1/R3-9 text
diagnoses its own recurring defect as a *citation-discipline* problem and proposes only the claim
manifest (§3 row 5) as the fix. That is a real but narrower fix — it would make each round's
footnote text more accurate, but it does nothing to make the contested-docket mechanism *detect*
that the same underlying dispute is recurring. Blue's own existing, related fix elsewhere in the
report (§2.1 Tier A, "Gap-id rollover across non-adjacent rounds... `prevGapIds` holds only the
prior round... widen `prevGapIds` to full adjudicated history") **would not close this either**: it
addresses a *same-id-recurring-later* case, not a *fresh-id-each-round* case — widening the
historical id set does not help when the successor gap was never given a matching id in the first
place. The backlog's own proposed fix (`supersedes: [prior-ids]` field on the gap envelope +
lineage-following contested-detection, i.e. escalate to the judge once a lineage chain reaches
depth ≥ 2, regardless of id identity) is a different and more targeted mechanism than either of
blue's two existing proposals, and the report does not mention it anywhere (full-report grep for
"lineage"/"supersedes"/"docket detector" returns zero hits outside this footnote's own history).

**Corroboration confidence: HIGH** — direct code trace (`debate.js:178`), direct `git log -p` read
of the live commit introducing the diagnosis, and direct cross-check of blue's own CHANGELOG
against the exact gap-ID chain the backlog names.

**Grading:** likelihood **certain** (this is not a projected risk — it already happened, in this
report, to this footnote, three consecutive rounds, and the backlog independently confirms it as a
"live run-3 discovery" rather than my own inference alone). Impact **high**: the contested-docket/
judge-adjudication path is this system's entire mechanism for distinguishing "red is legitimately
still finding new problems" from "red and blue are spinning on the same unresolved dispute" — and
that mechanism never engaged once for the single defect class that recurred most in this very
report; the only thing that stopped the cycle was the round budget, not adjudication. Complexity
**low-medium**: the fix is already scoped by the live source (add `supersedes: [prior-ids]` to the
gap envelope; change the contested-filter to follow lineage chains at depth ≥ 2; add a simulator
regression case) — this is closer to the "gap-id rollover" row's existing low-complexity grade than
to a redesign.

**Required fix:** add this as its own graded row (distinct from, and superseding as the primary
fix for, the existing "gap-id rollover across non-adjacent rounds" row, which should be kept but
marked as a narrower sub-case), cite `ideas/backlog.md`'s docket-detector item directly, and add a
§5 open question / §2.3 simulator case for the lineage-chain scenario. These land in §2/§3/§5,
outside this lens pass's assigned slice (§0–§1) — flagging here at the slice this finding's
evidentiary anchor lives in (`[^DiminishingReturns]`'s own footnote history), for the merge seat to
route into the correct sections.

## Not re-raised (checked, no new finding)

- CVE-2026-21852 mention in §1.1 (no footnote attached to this specific clause in my slice; it is
  inherited, already-corroborated fact from the memory-architecture run's own prior red audit, not
  a fresh citation introduced in this retrospective — out of this lens's leaf-node scope since
  there is no reference in this slice to follow).
- All `[^DiminishingReturns]`, `[^AgentDiversity]`, `[^NarrativeSimilarity]`, `[^IsolatedCorrection]`,
  `[^WisdomCrowds]`, `[^DiversityCollapse]`, `[^ProvenanceSurvey]` figure-level claims — no CHANGELOG
  change to §1.1 beyond R3-4/R3-9 (already ledgered HIGH/MEDIUM at their current confidence); not
  re-fetched per the ledger skip clause.
- `[^ResearchCommand]`, `[^Run2Frontier]`, `[^WorkflowJs]`, `[^PR14]`, `[^SimulatorTests]`,
  `[^LiveBacklog]`, `[^MainVsBranch]`, `[^Reverify47ae48d]`, `[^JudgeUnguarded]`,
  `[^PinnedRepoState]`, `[^CatechismTemplate]`, `[^R2-10]` — all HIGH in ledger, no section change
  this round, not re-fetched.

## Verdict (this slice only)

Section §0/§1 as written: no unrebutted low-confidence citations found this pass; the one new
finding (R4-1) is a structural/mechanism gap discovered by following a live citation forward, not a
sourcing-accuracy failure of any existing footnote. Recommend: **FAIL carries forward** pending
R4-1's disposition (this is a new, high-grade, unaddressed finding — not yet argued or
risk-accepted by blue).
