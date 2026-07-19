# red ledger — the single source of truth for gap status

Grading v2: `existence` (verified = checked at the leaf / suspected = inferred) is separate from
`likelihood`, which grades the CONSEQUENCE only. Closed prose lives in `archive.md` (append-only).

**Round 1 board:** 10 open, 0 closed. Max severity medium-high. Mass 37.0.
**Round 2 board:** 7 open (1 carried, 6 fresh), 9 closed. Max severity medium-high. Mass 25.5.
**Round 3 board:** 3 open (1 carried/re-raised, 2 fresh), 15 closed. Max severity medium-high. Mass 14.0.

**ROUND-3 CONDITION, READ THIS FIRST.** Blue took no round-3 turn. Verified: `records/` contains no
`events-blue-respond-r3-*.jsonl`; `blue/report.md` (00:49) and `blue/CHANGELOG.md` (00:50) both predate
judge-r2's sitting (00:57) and red's round-3 lenses (01:01). The artifact audited this round is
byte-identical to the one the bench ruled on. **Zero closures this round** — red closed nothing because
nothing was repaired, and the six round-2 dispositions were the bench's, not red's. R2-2's carried
obligation is unmet by non-response, not by a failed repair, and this round's FAIL should be read that
way by the stopping judgment.

---

## OPEN GAPS

### R2-2 — the parsimony disposal is still performed globally at the one site that performs it *(carried by judge-r2; re-raised unrepaired)*
- **class:** causal-overreach · **found_by:** [] (merge-only; L5-F2 re-surfaced it this round) · **supersedes:** ["R1-7","R1-2"]
- **location:** "Provenance and limitations of this round" — *"Both do not need to be true; the resolver account holds at the leaf."*
- **problem:** Two of three legs shipped in round 2 and the bench credited them: Catechism 3(b) and §2 now partition the causal claim by session type and concede that on the interactive branch "the mechanism producing empty blocks on that branch is unresolved and the serialization hypothesis remains live there." The third leg, at a site the pre-agreed check named, is untouched. Provenance lines 644-648 still dispose of the rival account over the whole corpus: *"the display-resolver finding … is the more parsimonious explanation (a single guard forces `display:"omitted"`) versus the serialization claim … The resolver path is implemented in code; serialization would require a second bug. Both do not need to be true; the resolver account holds at the leaf."* No partition, no carve-out. The report says both things: the hypothesis is live on the interactive share (two prominent sites) and retired over the corpus (the one site where the retirement is actually argued).
  Judge-r2 carried this gap and stated the obligation exhaustively at `debate.md` line 464: "one clause at the Provenance adjudication (report lines 644-648) limiting the parsimony disposal to the non-interactive share… If that clause lands, this closes." Verified at the leaf this round by `grep -n "parsimon\|holds at the leaf" blue/report.md` — two hits, both in Provenance, both unchanged. The clause has not landed because blue did not respond this round.
  Red re-raises rather than escalating. The bench's reasoning for carrying — that this residue runs toward *over*-claiming, unlike R2-1's, and that the archive shows the same non-propagating-edit failure twice already in this gap's own ancestry — is unchanged and red adopts it without addition.
- **required_fix:** Unchanged from the bench's statement, which red does not extend: one clause at Provenance lines 644-648 limiting the parsimony disposal to the non-interactive share and stating that the single-guard premise is unavailable on the interactive share, so the serialization hypothesis is not retired there. Nothing else is owed on this gap — no probe, no further partition, no re-argument of the causal claim.
- **acceptance_check:** DOCUMENT-PROBE — `grep -n "parsimon\|holds at the leaf" blue/report.md`; every hit must sit inside a sentence that names the session-type share it applies to, and no sentence may state the resolver account settles the serialization question over the whole corpus.
- **existence:** verified · **severity:** medium · **likelihood:** medium-high · **impact:** medium · **complexity_cost:** trivial
- **grade movement this round:** severity medium-high → medium; impact medium-high → medium; likelihood medium-high (unchanged); complexity low-medium → trivial. **Basis:** two of three legs shipped in round 2 and the remaining defect is one un-carved-out clause in a limitations section, not an unfenced causal claim at the headline — the prominent sites now carry the correct partition, which bounds how far a reader can be misled. Complexity is trivial on the bench's own finding that the fix is a single clause. Recorded via `regrade`.

### R3-1 — the Provenance section is the one site no repair has ever been swept against, and it now contradicts the body at three places
- **class:** incomplete-repair-lag · **found_by:** ["L5"] · **supersedes:** ["R2-1","R2-5"]
- **location:** "Provenance and limitations of this round" — *"Not verified this round: … the arXiv identifiers in [^AgentBenches], **the NIST quotation's primary source**, and the vendor sales channel for raw thinking."*
- **problem:** LINEAGE DECLARED: this gap amends two closures judge-r2 entered in round 2 (R2-1, R2-5), both read in full at `red/archive.md` and at `debate.md` before minting. It does not re-litigate either ruling; it says the two flagged residues plus R2-2's unmet leg are three instances of one defect at one site, and that the third instance is new business no seat has seen.
  (i) **New business — the not-verified list contradicts its own footnote.** Provenance line 656 lists "the NIST quotation's primary source" among what was **not verified this round**. The NIST quotation was retired in round 2 and `[^NISTAuditRequirement]` (line 623) now records "Secondary source; verified 2026-07-19". The report's limitations list disclaims the primary source of a quotation the report no longer contains, and states unverified what its own footnote states verified. Judge-r2's R2-5 residue note covers the footnote **label** only ("invisible in rendered output and outside the acceptance check"); this is a different site, is visible in rendered output, and runs in the direction the bench's economy principle does not protect — a reader is told a live footnote is unverified when blue verified it.
  (ii) **Amended from R2-1's flagged residue.** Provenance lines 642-644 still say the independence claim "is retracted provisionally **pending confirmation** whether these are the same measurement or independent sweeps" while §2 line 191 states the resolution and blue's CHANGELOG claims the measurements "match exactly". The bench flagged this not-ruled and declined to carry it alone, on the ground that it errs toward under-claiming. Red accepts that grading in isolation and does not dispute it. What the bench could not see, ruling gap-by-gap, is that this is the *second* stale statement in a section that also carries the *third* (R2-2's clause) and the *first* (leg i).
  (iii) **The class.** Three rounds of repairs have each been verified against the sites the fix-spec named and have each left this section stale, because Provenance is where every round's honest hedges were parked and no repair has ever been swept against it. Blue's round-1 and round-2 "corrections propagated report-wide" lists name no Provenance site. That is the defect: not any one sentence, but that the section documenting the report's limitations is structurally outside the report's own propagation discipline. It is also the section a skeptic reads to check how competing accounts were handled.
- **required_fix:** Reconcile the Provenance section against the current body, and add the section to the standing propagation checklist. Specifically for the two legs red names: strike "the NIST quotation's primary source" from the not-verified list (or restore the claim it disclaims); and state the sweep-independence question at the same confidence Provenance and §2 both use — resolved, or open with the outstanding confirmation named, at both sites. **CLASS RULE:** the Provenance/limitations section is a propagation site like any other; every future repair's site sweep must include it, and any hedge, retraction, or not-verified entry parked there must be re-checked against the body whenever the claim it covers is edited. The enumeration of stale Provenance statements is declared **OPEN** — these are the ones red found, not a closed list, and the fix is the sweep, not the two edits.
- **acceptance_check:** DOCUMENT-PROBE — (a) `grep -n "NIST" blue/report.md`: no hit in the not-verified list may name a NIST quotation the report does not contain, and no hit may state unverified what `[^NISTAuditRequirement]` states verified; (b) `grep -n "pending confirmation\|provisionally" blue/report.md`: no hit may describe the 287/5,569 question as open while §2 describes it as answered; (c) blue's round-3 CHANGELOG propagation list must name the Provenance section, which is how red checks the class rule landed rather than the two instances.
- **existence:** verified · **severity:** medium-high · **likelihood:** medium-high · **impact:** medium · **complexity_cost:** trivial

### R3-2 — the session-type partition is unquantified, and its two numerals collide at 278 by coincidence
- **class:** numeric-collision-under-partition · **found_by:** [] (merge-only)
- **location:** §2 "The mechanism, read out of the shipped client" — *"The report's own §1 counts 16 top-level transcripts (interactive parent sessions) out of 294 files, meaning 278 are deeper-nested subagent and workflow runs."*
- **problem:** This sentence is new in the R2-2 repair, and both the bench and red credited that repair without checking its arithmetic. NO SUPERSEDES: R2-2 is open and carried, so this cannot supersede it; it is fresh business arising from the same repair, and red records the adjacency here rather than in a lineage field it is not entitled to use.
  (i) **Two different 278s, twenty-seven lines apart in the same section.** Line 185: *"found 294 transcript files, **278** of which contain thinking blocks"* — that 278 is `grep -l '"type":"thinking"'` per `[^LocalSweep]`. Line 213's 278 is 294 − 16, the count of *nested* files. Different sets, equal cardinality, and the partition reads as covering the corpus only because the numerals match. The report's own quoted evidence proves they are not the same set: the pinned probe, quoted at Provenance line 640, reports the empty blocks are *"Consistent across seat and **main-session** transcripts"* — main sessions are the top-level ones, so at least one of the 16 carries thinking blocks, so strictly fewer than 278 nested files do. The consequence: the report never states how many of the 5,754 blocks fall on the interactive share whose mechanism it has just conceded is unresolved. It could be a handful or a large minority; the partition the bench credited is stated but not sized, and a reader will size the unexplained share at zero from the matching numerals.
  (ii) **Nesting depth is asserted to be session type, and §1 does not say so.** §1 says only *"a top-level glob of the projects directory found 16 files where a recursive walk found 294"* — a filesystem fact. §2 attributes to §1 a characterization §1 does not make ("16 top-level transcripts (**interactive parent sessions**)") and makes that equation load-bearing for which branch of the display resolver applies. Whether directory depth tracks `isNonInteractiveSession` is unverified, and it is exactly the discriminator the partition needs.
- **required_fix:** Either (a) partition the block count by the property that decides the branch — re-run the sweep grouped by session type, or by the depth proxy with the proxy's soundness argued — and state how many of the 5,754 blocks sit on each share; or (b) state plainly that the split is unquantified, that the 278-with-thinking and 278-nested figures are distinct counts that coincide, and that the interactive share's block count is unmeasured. In either case stop attributing the "interactive parent sessions" gloss to §1, which does not make it. **CLASS RULE:** every figure reused across sections must be checked for set identity, not numeral identity, before an inference is drawn from the match; and any proxy standing in for the property a claim turns on (here: nesting depth for session type) must be named as a proxy with its soundness argued at the site of use. Enumeration open.
- **acceptance_check:** DOCUMENT-PROBE — read §2 lines 184-215. The two 278s must be distinguished in the text (or one replaced by a measured figure), the interactive share's block count must be given or explicitly declared unmeasured, and the "interactive parent sessions" gloss must be dropped, sourced to a stated depth-to-session-type argument, or attributed to something other than §1. A LIVE-PROBE (re-running the store sweep grouped by session type) would discharge option (a) but is **not demanded**: option (b) is a document edit and is a complete answer.
- **existence:** verified · **severity:** medium · **likelihood:** medium · **impact:** medium · **complexity_cost:** low

---

## CLOSURE INDEX

*(id | closure class | one-line summary | supersedes)* — 15 lines. R1-1…R1-8 and R1-10 closed by
red-merge-r2; R1-9 and R2-1/R2-3/R2-4/R2-5/R2-6 closed by judge-r2 at the round-2 sitting, with red's
independent leaf confirmation of each recorded in `archive.md`. **Zero closures were entered in round 3.**

| id | class | summary | supersedes |
|---|---|---|---|
| R1-1 | closed_with_regression | 0.417 re-cited to arXiv:2603.05488 with 0.012 and task-dependence; the repair introduced a generalization the paper's second model arm refutes | — |
| R1-2 | closed_with_regression | Pinned inputs stated recoverable at `cacb736`, both read, serialization hypothesis quoted and adjudicated; "independent" left standing at §2 and the adjudication does not cover the interactive share | — |
| R1-3 | closed | `feov-record blue --help` re-run at merge; footnote's verb list matches the tool output exactly and red-seat verbs are attributed to red | — |
| R1-4 | closed_with_regression | dev.to citation retired and truncation re-grounded on binary evidence; the `[^ToolTruncation]` reference marker was left behind at the Catechism | — |
| R1-5 | closed_with_regression | meta-intelligence.tech retired and §8 de-dated; the substituted zylos.ai source carries neither the quotation nor the NIST attribution, and "Q4 2026" survives at open question 7 | — |
| R1-6 | closed_with_regression | 260+ replaced by ~30 with the conflict disclosed; the replacement figure is attributed to two sources that do not carry it | — |
| R1-7 | closed_with_regression | Scope conditions added inline at headline and Catechism 3; cause-vs-consistency made explicit — and the newly explicit causal claim overreaches its own evidence base | — |
| R1-8 | closed | §4 hedged to "on this version"; report-wide sweep for ever/never/always found no remaining absolute attached to a binary-derived finding | — |
| R1-9 | closed (bench, round 2) | §8 grades adjudication-time verification as equal-to-higher per claim ("even if costlier per claim"), distinct from the recording- and maintenance-cost sentences | — |
| R1-10 | closed | §6 composition rule added — a claim spanning tiers grades at its weakest leg, with the legs named and an example | — |
| R2-1 | closed (bench, round 2; residue flagged not ruled) | "independent" struck at §2 point of use; Provenance retraction still framed as pending — the flagged residue is amended by R3-1 | R1-2 |
| R2-3 | closed (bench, round 2) | Both model rows at §2 and `[^ReasoningTheater]` with figures pinned to DeepSeek-R1 671B by name; Table 1 re-fetched live at red-merge-r3, no drift | R1-1 |
| R2-4 | closed (bench, round 2) | Catechism marker repointed to `[^ToolTruncationLimits]`; class re-audited at red-merge-r3 — zero orphans in both directions for both retired labels | R1-4 |
| R2-5 | closed (bench, round 2; cosmetic residue noted) | NIST attribution and absent quotation dropped from footnote and §8, "Q4 2026" struck; the not-verified-list contradiction is new business, carried by R3-1 | R1-5 |
| R2-6 | closed (bench, round 2) | ~30 now cited to generalanalysis.com, the two navigation pages labelled non-carrying, parenthetical restated as a count conflict | R1-6 |

---

## DECLINED LENS OBSERVATIONS (not gaps; recorded so they are not re-litigated silently)

### Round 1 (carried)

| Observation | Fate | Reason |
|---|---|---|
| L5-F5 (omitted as privacy-by-design) | declined | Bears on desirability, not on the capability question the report asks. |
| L5-F7 (risk matrix omits the recommendation's failure mode) | declined | Refuted by direct read: §9 row 6 is that row. |
| L5-F9 (future default-flip counterfactual) | declined | Carried at §9 row 8 and §7 stopping point (i). |
| L6-F3 r1 (sweep discipline unsystematized) | declined, risk-argued | Scheduled-sweep machinery is complexity that makes the design worse for a disclosed drift. |
| L6-F4 r1 (JSONL not tamper-evident) | declined | Vendor property outside blue's control, already stated at §1. |
| L6-F7 r1 (metric-name drift) | declined | Corrected in §4 and risk-accepted at §9 row 4. |
| L1 r1 (CLEAN over §§1–5) | banked | Merge found two defects inside that slice. Recall signal against L1. |

### Round 2 (carried)

| Observation | Fate | Reason |
|---|---|---|
| L1-F1 (performativity condition misattribution) | minted-as R2-3 | Adopted and sharpened by the merge's own Table 1 fetch. |
| L1 obs#1 (store grew 294→306) | declined | Moving-target property disclosed in-line and at `[^LocalSweep]`. |
| L1 obs#2 (Compliance API docs 404) | folded-into R2-6 | Lens fetched a mistyped path; the live defect was different and worse. |
| L5-F1 (broken `[^ToolTruncation]` reference) | minted-as R2-4 | Confirmed by grep. |
| L5-F2 (Q4 2026 in open question 7) | folded-into R2-5 | Same R1-5 lineage. |
| L5-F3 (permission decisions misplaced in Tier 1) | declined | Refuted by direct read: `vc("tool_decision", …)` is a first-class event. |
| L6-F1 (store volatility) | declined | Disclosed at point of use and at `[^LocalSweep]`. |
| L6-F2, L6-F16 (mechanism never empirically validated) | folded-into R2-2 | The document-checkable form was carried instead. |
| L6-F3 (falsifying experiment declined) | declined | Settings mutation outside seat consent; the substitute demand shipped at R1-7. |
| L6-F4, L6-F12 (artifacts conflate recording with reasoning quality) | declined | Conceded in blue's own voice at §8, the case-against, and §6 Tier 4. |
| L6-F5 (truncation disclosure unenforceable) | declined, risk-argued | Already the report's position at §9 row 3; the control is worse than the priced risk. |
| L6-F6 (version-bound shelf life) | declined | Stated at §4, §9 row 8, Provenance. |
| L6-F7 (absence claim bounded) | declined | Fenced at §3 and `[^ComplianceAPI]`. |
| L6-F8 ("platform explicitly forbids") | declined, out of surface | Targets `lines-of-inquiry.md`, not the report. |
| L6-F9 (future export paths) | declined | Speculative future-version risk blue does not control. |
| L6-F10 (visibility gap N=1) | declined | §5 labels it a single anecdotal report and declines the causal claim. |
| L6-F11 (adaptive thinking effort) | declined | Carried at §3 and §7 item 7; mitigation is a vendor API change. |
| L6-F13 (recursive globbing) | declined | Stated at §1 in the report's own numbers. |
| L6-F14 (re-enumeration not trivial) | declined, risk-argued | Grade dispute on a risk already accepted. |
| L6-F15 (artifact effectiveness unquantified) | declined | Figures fenced as unverified with only the direction carried. |
| L6 summary self-inconsistency | banked | Lens summary cited labels its own pass never wrote. |

### Round 3

| Observation | Fate | Reason |
|---|---|---|
| L1 (CLEAN over Catechism + §§1–5, 16 citations) | banked | The artifact was unchanged from the round the bench already ruled on, so a CLEAN pass carries little information — but R3-2 sits at §2 lines 185/213, inside L1's declared slice, and is arithmetic rather than citation. L1 has now returned CLEAN over a slice containing merge-found defects in two of three rounds. L1's own note that it "RELAYED" nine round-0 verifications without re-checking is the mechanism, and its coverage model verifies citations rather than reading the section. |
| L2 (CLEAN over §§6–10, open questions, Provenance) | banked, recall failure | L2's slice **is** the Provenance section. It reported "All citations in my slice verified" and "No defects detected" while that section carried the bench's own carried obligation (R2-2), a bench-flagged residue (R2-1), and an unnoticed contradiction of its own footnote (R3-1 leg i). L2 checked citation reachability and never read the section against the document it limits. The round's clearest recall miss, and it is against the one section the bench had just told the board to look at. |
| L5-F2 (resolver-vs-serialization not closed) | folded-into R2-2, credited on R3-1 | Correct and correctly anchored — the lens quoted the exact Provenance sentence the bench carried. It is R2-2's unmet leg, not a fresh gap, so it is not minted twice; L5 is credited as found_by on R3-1, whose class the observation established. |
| L5-F1 / L6-F4 (escalate the `showThinkingSummaries` test to the operator) | declined | Third consecutive round. The settings mutation is outside seat consent, the dependency is named at open question 1 and §7 stopping point (i), and the headline carries the condition the experiment would test. L5's framing — escalate rather than silently defer — is an operator-practice recommendation and is passed to the lead in red's position rather than minted as a report defect. |
| L5-F3 (closed issues stale relative to v2.1.215) | declined | Refuted by direct read: the same case-against bullet that spends the closure-status evidence carries the staleness caveat in its own final clause — "our account of current behavior rests on a locked thread describing v2.1.71, two hundred patch versions stale". The lens asked for a caveat already in the sentence it quoted. |
| L5-F4, L5-F9 (tier-discipline complexity underestimated) | declined, risk-argued | A grade dispute on one risk-matrix cell. The row prices labelling *trajectory claims about agents*, not this report's prose; and the two figures cited as evidence of laxity (IBM 45/94, 500K `maxResultSizeChars`) are labelled unverified at the point of use, at the footnote, and again in the not-verified list — the tier discipline working, not failing. |
| L5-F5 (faithfulness case rests on other models) | declined | Refuted by direct read: `[^ReasoningTheater]` states "**Not** generalized to Claude models in this report", and the Claude-specific leg of the headline runs through Anthropic's own documentation declining to warrant summaries as faithful — the vendor on its own model, labelled as such. |
| L5-F6 (artifact recording timing not stated) | declined | The post-hoc-rationalization risk is named at §9 row 6 and risk-accepted with rationale; §8 concedes the channel buys "not sincerity". Real-time-vs-end-of-run enforcement is a property of the recording tool, not of this report's finding. |
| L5-F7 (adaptive thinking conflates effort with adjudication) | declined | Refuted by direct read: the §3 sentence the lens quotes already scopes itself — "making **reasoning-quality** adjudication impossible without controlled re-execution". |
| L5-F8, L6-F6 (Compliance API bounded to public surface) | declined | Third round. Fenced at §3 in the same sentence, at `[^ComplianceAPI]`, and carried as open question 10. |
| L6-F1 (false artifacts undetectable) | declined | Blue's own position at §8, in the case-against, and at §9 row 6 where it is priced medium/medium/high and risk-accepted. The demanded detection mechanism is the unsolved problem the report names, not a defect in naming it. |
| L6-F2 (cost model inverts at adjudication) | declined, closed elsewhere | This is R1-9, which the bench closed: §8 now concedes the inversion the lens asks for — "even if costlier per claim than a thinking-block read alone". The lens read the report without the paragraph. |
| L6-F3 (truncation detection rule) | declined, risk-argued | Unchanged from round 2. |
| L6-F5 (version-binding has no automated safeguard) | declined, risk-argued | Unchanged from rounds 1 and 2. |
