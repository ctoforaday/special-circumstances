# Round 3, Lens 1 — leaf-node citation verification, slice: §0 (Headline) + §1 (Doubts: H1/H2/H3/un-frontiered)

Full living `blue/report.md` (770 lines) re-read in full this pass, not just the CHANGELOG diff.
Ledger skip-clause honored: claims already HIGH in `red/citation-ledger.md` whose section was not
listed as changed in `blue/CHANGELOG.md`'s Round 2 entry were not re-fetched. Re-verified live:
round-2 changes touching this slice (R2-2 §0 addendum, R2-3 §1.1 cross-provider paragraph, the
R2-4 footnote spillover into §1.1 body), plus the current `main` HEAD and the red-auditor memory
store (both directly inspectable, not requiring web fetch).

## Verdict for this slice: FAIL — 1 new gap (R3-1), everything else in §0/§1 verified clean or
correctly closed.

## R3-1 — OPEN — MEDIUM — certain x medium x trivial — corroboration: HIGH (direct text
comparison within the shipped document; footnote's replacement claim independently re-verified
against arXiv:2606.02646 full HTML)

**Location:** §1.1 (H1), disconfirming-evidence bullet on diminishing returns — *"with the
breakeven shifting to 3–4 agents on harder tasks and continued gains observed to 7 agents on the
hardest — so the qualitative thesis holds; the precise '2–4' is a synthesis across sources, not a
single citable number.[^DiminishingReturns]"*

**Problem: body prose still asserts the exact clause its own footnote retracted this round —
repair-regression continuing one hop further, footnote-fixed/body-stale.** Round 2's R2-4 rebuttal
(`blue/CHANGELOG.md` Round 2, and the footnote `[^DiminishingReturns]` itself) directly fetched
red's proposed source (arXiv:2606.02646) and found it does **not** support "continued gains to 7
agents on the hardest" — re-verified independently this round (WebFetch of both the abstract page
and `arxiv.org/html/2606.02646v1`): the paper's hardest benchmark is GSM-Hard (1,017 free-form
numeric word problems), the practical knee for harder tasks is **N≈10**, and its headline finding
is that effective team size **plateaus around 1.8 agents by N=30**, with a single N≤5 pilot
predicting that ceiling — i.e., the paper's actual direction is *earlier*, not later, diminishing
returns on hard tasks. Blue's own footnote correctly states this and explicitly says the "7 agents"
clause is "dropped rather than re-cited to an unverified source," restating the synthesis "without
a specific agent-count ceiling for the hardest tier" and noting the corrected direction is "toward
*more* caution about adding agents on hard tasks, not less."

But the §1.1 **body** paragraph carrying this claim — written in round 1 (R1-5) and never touched
by the round-2 footnote fix — still reads "continued gains observed to 7 agents on the hardest,"
presented as corroborated by "independent re-search this round" and footnoted to the very note that
now says the opposite. A reader of the body prose alone (the overwhelming majority of the text; the
footnote is 400+ words down in a footnote block) takes away the retracted, wrong-direction claim.
This is the same failure class as R1-5→R2-4 itself (an uncited/miscited precise figure surviving
inside its own "fix"), recurring one layer further out: the fix landed in the footnote's prose but
not in the sentence the footnote is attached to. Distinguish from the memory pattern of *footnote
lagging body* — here it is the reverse, body lagging the (correctly repaired) footnote.

**Required fix:** edit the §1.1 body clause to match the footnote's own corrected synthesis — drop
"continued gains observed to 7 agents on the hardest" and replace with wording consistent with the
footnote (e.g. "with the breakeven shifting higher on harder tasks — though per the report's own
corrected footnote, effective diversity may saturate *earlier*, not later, on the hardest tasks").
Trivial edit; no new research needed, the correct direction is already sitting in the footnote two
sentences away.

## Verified clean this slice (leaf-node re-checked or ledger-confirmed unchanged-section)

- **R2-2 fix (§0 live addendum):** "across rounds 0 and 1" now used consistently in all three
  places (§0 addendum, inherited phrasing); no residual "this same round" contradiction found on
  full re-read. CLOSED, corroboration HIGH.
- **R2-3 fix (§1.1 cross-provider paragraph):** now correctly attributes "2 vs. 16" to the paper's
  L4 (model+persona) condition and states L2's real curve (8 agents to match L1@16, 65.44% vs.
  65.34%); defer-not-adopt disposition now rests on the infrastructure-cost argument alone. Matches
  ledger's HIGH-confidence Table 2 re-fetch (citation-ledger.md line 86). CLOSED, corroboration
  HIGH.
- **R2-4 footnote itself** (as distinct from the body clause it's attached to, above): the
  rebuttal's factual claims about arXiv:2606.02646 — GSM-Hard (not GSM-Plus), "practical knee is
  N≈10," "effective team size plateaus ~1.8 agents by N=30," "single N≤5 pilot predicts the N=30
  ceiling" — independently re-verified this round via direct fetch of the full HTML text (v1). All
  four figures/claims confirmed exact. CLOSED, corroboration HIGH.
- **§1.4 red-auditor memory-loop claim** — "only `red-auditor.md` declares `memory: project`; blue
  and judge do not" — independently re-verified this round by reading
  `plugins/frank-exchange-of-views/agents/blue-researcher.md` and `.../lead-judge.md` frontmatter
  directly: neither has a `memory:` key (only `name`/`description`/`tools`/`skills`). Confirmed
  true. Corroboration HIGH (new leaf-node check this round, not previously ledgered as a negative
  claim).
- **§1.4 "15 well-formed gap-pattern files" / "generalizing R4-1's denylist-vs-allowlist finding
  into 'prove incompleteness via the system's own symmetric defense; recommend allowlist
  inversion'"** — the descriptive quote matches `pattern_invariant_soundness_by_enumeration.md`
  verbatim in substance (confirmed by direct read). The "15" count itself is now further stale
  (live store holds 23 pattern files, up from 18 at round-2 audit time) — **not re-raised**: this
  exact drift was already triaged and dispositioned as "noted, not raised" in round 2
  (`red/findings.md` line 351, "stale-count churn inherent to describing a live accreting
  mechanism") and nothing has changed about that disposition; re-litigating a settled, correctly-
  reasoned non-gap every round would violate the "don't grind through more rounds" norm. Logged in
  the ledger for the running count only.
- **Live-source drift, no impact:** `origin/main` has advanced again, `88eb57f` → `d164ab2`
  (docs-only backlog commit, 1-line diff in `ideas/backlog.md`, confirmed via `git diff --stat`).
  No change to `debate.js` or any file this slice's claims depend on. Same pattern as the
  round-2-noted `47ae48d`→`88eb57f` drift; [^PinnedRepoState]'s discipline continues to work as
  designed. Not a gap — logged for the pattern.
- All other §0/§1 citations (PR #14 merge state, judge-unguarded, citationPasses-const,
  arXiv:2604.18005 DiversityCollapse, WisdomCrowds, arXiv:2605.00914 IsolatedCorrection,
  arXiv:2606.04990 ProvenanceSurvey, ClaimManifest backlog item, CatechismTemplate diff,
  LocalGrep/LocalGrepRed/BlueReportGrep/BlueReportUnverified/RedFindingsGrep/RedAuditorSpec) were
  already HIGH in `red/citation-ledger.md` from rounds 1–2 and their sections were not listed as
  changed in `blue/CHANGELOG.md`'s Round 2 entry — not re-fetched, per protocol.

## Ledger appends (this pass)

- `plugins/frank-exchange-of-views/agents/blue-researcher.md` and `.../lead-judge.md` frontmatter
  have no `memory:` key (only `red-auditor.md` does) | direct read, this round | HIGH | round 3
- arXiv:2606.02646 states GSM-Hard (not GSM-Plus), "practical knee is N≈10" on harder tasks,
  effective team size plateaus ~1.8 agents by N=30, single N≤5 pilot predicts N=30 ceiling |
  `arxiv.org/html/2606.02646v1`, live fetch this round | HIGH | round 3
- §1.1 body "continued gains observed to 7 agents on the hardest" contradicts its own footnote
  [^DiminishingReturns]'s round-2 correction (which drops this exact clause and states the opposite
  direction) | direct text comparison, blue/report.md, this round | HIGH (gap exists) | round 3
- `origin/main` advanced `88eb57f` → `d164ab2`, docs-only (`ideas/backlog.md` 1-line), no
  `debate.js` diff | `git diff 88eb57f d164ab2 --stat`, live, this round | HIGH | round 3
- red-auditor memory store now holds 23 gap-pattern files (was 18 at round-2 audit) | `ls`
  `AgentOrange/.claude/agent-memory/frank-exchange-of-views-red-auditor/`, this round | HIGH
  (fact) — not re-raised as gap, already dispositioned round 2 | round 3
