# Red audit — round 2, lens: dark-side and risk

Scope: failure modes, likelihood x impact x complexity grading, security and tradeoff blindspots.
Full re-read of `blue/report.md` (704 lines, post-Round-1-corrections) completed in context, not
from `blue/CHANGELOG.md`'s Round 1 diff summary alone. Leaf-node verification against the live
`special-circumstances` repo: `git fetch origin` shows `main` has advanced again, to `88eb57f`
("docs(backlog): run cost audit"), one commit past the report's own pinned `47ae48d` — checked the
diff (`ideas/backlog.md`, +1 line, docs-only) to confirm it does not touch `debate.js` and does not
invalidate any of the report's code-level claims; direct full read of `debate.js` at `origin/main`
confirms the judge-null (row 2), `citationPasses`-const (row 2b), and bulk/judgment routing (row
16b) claims all still hold exactly as the report states. Also checked `agents/red-auditor.md`,
`skills/research-protocol/SKILL.md`, and `commands/research.md` at `origin/main` for the actual
mechanics behind three of the report's own risk-accept rationales.

## Verdict: FAIL — 4 new gaps (G1–G4). None overturn H1–H5; three are real, previously-unexamined
dark-side findings on material that entered the report *during* Round 1 (the redundancy floor, the
poisoning risk-accept, the shipped citation ledger); one is a smaller cross-cutting observation.
Round 1's 20 gaps (R1-1..R1-20) are not re-litigated here except where this lens found the fix
itself introduces a new risk — see the disconfirming pass for what was checked and held.

---

### G1 — The poisoning risk-accept's named mitigation claims a capability the protocol doesn't actually grant, and the report's own next paragraph half-admits it [MEDIUM likelihood x HIGH impact x LOW-MEDIUM complexity]

**Location:** §3 row 19 disposition — *"a poisoned page asserting a fake fact still has to survive
**independent re-verification against a second source**."* Compare §5, item 8 — *"Is a poisoning
attack against the citation itself ... covered by the leaf-node citation-verification lens, or does
it require **a distinct defense (e.g. cross-referencing claimed sources against an independent
index)**? ... not resolved here."*

**Problem:** row 19 risk-accepts the reflexivity gap (FEOV's own WebFetch/WebSearch-driven research
phase as an untrusted-content-poisoning surface) on the strength of a named mitigation — "independent
re-verification against a second source." But the actual protocol text this lens operates under
(`research-protocol/SKILL.md`: *"AFTER drafting, every claim MUST trace to a source a skeptic can
follow"*; `red-auditor.md`: *"During the audit, YOU MUST verify claims at the leaf node: **follow
the citation to the source**; confirm the source actually corroborates the statement"*) describes
re-fetching the *same* cited URL and checking it says what the footnote claims — not searching for
and cross-checking against an independent second source. The word "independent" does not appear
anywhere in the `frank-exchange-of-views` plugin source (`git grep -n independent` — zero hits).
If a single page is poisoned (compromised, or contains a prompt-injection payload consistent with
its own fabricated claim), blue's draft and red's "verification" both consult the *identical*
untrusted channel — that is not independent corroboration, it is the same source checking itself
twice. The report's own §5 item 8, added in the *same* Round 1 revision as row 19, effectively
concedes this: it explicitly names "cross-referencing claimed sources against an independent index"
as a *hypothetical, not-yet-built* distinct defense, i.e. acknowledges the current lens does not do
this — directly contradicting row 19's claim that the mitigation already requires surviving a
"second source." One of these two round-1-added passages is wrong about what the leaf-node lens
currently does; as written, they disagree with each other on the same page. This matters because
row 19 is a *closed* risk-accept ("Risk-accept with a named, disposed mitigation, not silence") —
its closure rests on an overstated mechanism, so the residual risk it accepts is understated, not
fully disposed.

**Required fix:** either (a) correct row 19 to state the honest mechanism — "the leaf-node lens
re-checks the cited source itself; it does not currently cross-reference an independent second
source, which is exactly the gap §5 item 8 already names as open" — downgrading the risk-accept's
confidence accordingly, or (b) add the missing teeth: require a second, independently-found source
before grading a load-bearing claim's corroboration HIGH when it currently rests on exactly one
external citation. Either closes the internal contradiction; leaving both sentences as-is does not.

**Corroboration confidence:** high (direct read of `red-auditor.md`/`SKILL.md`'s actual verification
instruction plus a repo-wide grep for "independent" — zero hits — against the report's own two
contradicting sentences, both quoted above).

### G2 — The redundancy floor added to fix R1-16 doesn't arithmetically fit inside the lane-count floor added to fix H1's under-provisioning; the report ships both without reconciling them [MEDIUM x MEDIUM x LOW]

**Location:** §3 row 6 — *"assign distinct method/source-class lenses (primary-literature /
practitioner-production / adversarial-disconfirming-first / local-repo critical-stance) ... WITH an
explicit redundancy floor: assign the critical-stance/adversarial lens to at least 2 of N lanes (not
1-of-N)"*; §3 row 7 — *"Lane-count floor (`lanes >= 3` or explicit justified override)"*.

**Problem:** row 6's fix (added Round 0, redundancy floor added Round 1 per R1-16) names at least
three distinct, simultaneously-desired method-classes — primary-literature, practitioner-production,
and adversarial/critical-stance (collapsing the two adversarial-sounding items into one, the most
charitable reading) — and requires the last of those to occupy *at least 2* lanes, not 1. That is a
minimum of 1 + 1 + 2 = **4** lane-assignments for full method coverage with the stated redundancy.
Row 7, added independently to fix H1's separate under-provisioning finding, floors lane count at
**3** — the exact default already in place (`lanes = 3` in `debate.js`) and the number the report
elsewhere treats as sufficient. At the stated floor of 3, the two adopted fixes cannot both hold as
written: either one method-class gets dropped/merged silently, or the redundancy floor quietly
becomes 1-of-N again for at least one run in three (whichever ran at the floor). Neither §3 row 6
nor row 7 does this arithmetic or cross-references the other row; each reads as though it is the
only constraint on lane count. This is the same class of gap R1-16 itself raised one level up
(an undiscussed interaction between two proposals that individually look fine) — now surfacing
between two of the fixes *this* report added to close R1-16 and the H1 lane-floor finding.

**Required fix:** one sentence reconciling them — either raise item 7's floor to `lanes >= 4` when
lens-assignment (item 6) is active, or state explicitly that at the 3-lane floor the redundancy
requirement is dropped (primary-lit and practitioner-production share a lane, or the redundancy
floor applies only above 3 lanes) — a stated tradeoff, not a silent one.

**Corroboration confidence:** high (both cited rows read directly from the current `blue/report.md`;
the arithmetic is internal to the report's own two sentences, no external source needed).

### G3 — The shipped citation ledger's own caching rule works against item 10's rationale for why live-source drift is an acceptable-cost risk, and the report never notices the two are now the same code path [MEDIUM-HIGH x MEDIUM-HIGH x LOW]

**Location:** §3 row 10 — *"Access-date-delta recording for citations ... Medium — drift is usually
caught by re-verification; the cost is re-work, not silent error."*; §0 — *"per-role model routing,
**the citation ledger**, the pre-created blackboard skeleton, and the Catechism template are now all
present on `main`."*

**Problem:** row 10's Impact-cell rationale for grading live-source drift only "Medium" (not High)
rests on the claim that drift "is usually caught by re-verification." But the very citation ledger
the report elsewhere celebrates as merged (§0, §3 row 1, §2.2) is live code on `main` today with the
opposite effect on exactly this axis: `debate.js`'s ledger clause instructs every red citation-lens
pass, "a claim verified at HIGH confidence in a prior round stays verified — do not re-fetch it
unless [blue's] CHANGELOG shows its section changed this round" (verified live, this round, direct
read of `debate.js`). That skip-condition is keyed to whether the *citing prose* changed, not to
whether the *external source* changed or how much access-date has elapsed — precisely the two
things row 10's own fix (access-date-delta recording) is meant to track. The corpus's own evidence
(mem0's ADD-only pivot, gh issue-status flips, star-count drift — all cited elsewhere in this same
report) shows external sources changing on the timescale of hours to a few days, well within a
single multi-round debate's runtime. Once a claim is graded HIGH in round 1, the ledger tells every
subsequent round's citation lens to skip it — so "re-verification" (row 10's stated safety net) is
exactly the step the shipped ledger is designed to *avoid* for any already-HIGH claim, for the rest
of the debate. The two mechanisms — the drift-mitigation the report recommends (row 10) and the
cost-saving cache the report has already shipped (the ledger) — now sit in direct tension in the
same file, and neither §0 (which only praises the ledger) nor row 10 (which only asks for
access-date footnotes) mentions the other.

**Required fix:** add a time- or access-date-based re-verification trigger to the ledger clause
itself (e.g., "...unless the section changed *or* N rounds have elapsed since the claim was last
verified") so item 10's access-date recording actually has a consumer that acts on it, rather than
producing a date field nothing re-checks.

**Corroboration confidence:** high (direct live read of `debate.js`'s ledger clause, quoted verbatim
above, against the report's own two cited sentences — both already in `blue/report.md`).

### G4 — Cross-cutting: several risk-accepts close on a "revisit if it recurs N times" trigger with no shared place that counts recurrences [LOW-MEDIUM x LOW-MEDIUM x LOW]

**Location:** §3 row 14 — *"Risk-accept with a revisit trigger ... Revisit if it becomes load-bearing
again"*; §3 row 15 — *"Risk-accept for run 4 ... track as re-graded to High likelihood; build the
chunking helper if it recurs a 4th time rather than continuing to absorb the workaround cost
indefinitely"*; §5, item 10 — *"Does ENAMETOOLONG recur a fourth time before the chunked-append
helper is built? ... a fourth occurrence should trigger the build, not another risk-accept renewal."*

**Problem:** three separate closures in this report lean on a future human (or a future retrospective
like this one) noticing and correctly counting a recurrence across runs — the exact kind of
archaeology this same report's own §4 counting-method note flags as unreliable ("35" vs. 21 friction
entries; "r1–r2" vs. the one instance the source actually attests), and the exact gap item 11
(trajectory capture) is proposed to close for defects generally. Unlike citations (which now have
`red/citation-ledger.md` as a durable, appended-to counter), these recurrence triggers have no
equivalent artifact — "the 4th time" is tracked only by whoever happens to remember and reread
`debate.md`/`CHANGELOG.md` across runs, the identical failure mode R1-7/R1-8 already caught for
friction-entry counts this same round. This is not a new finding in kind (it's the same
under-instrumentation the report already grades as a top priority via item 11), but it is a new
*instance*: two live risk-accepts (rows 14, 15) currently depend on a counter that item 11 has not
yet been built to provide, and the report doesn't connect that dependency explicitly.

**Required fix:** one added clause on item 11 (trajectory capture) or a new one-line note on rows 14
and 15: "this risk-accept's revisit trigger has no counter until item 11 ships; until then, treat the
trigger as advisory, not enforced."

**Corroboration confidence:** medium — reasoned from the report's own already-stated counting
failures (R1-7/R1-8) applied to a new instance (defect recurrence rather than friction-entry counts),
not an independently observed incident of a missed trigger.

---

## Disconfirming pass against this lens's own findings

- G1: could "follow the citation to the source" be read, in practice, as agents naturally also
  running a corroborating search rather than only re-fetching the one URL (i.e., the gap is
  documentation-only, and lens instances already behave more defensively than the letter of the
  protocol)? Checked `red/citation-ledger.md` and prior rounds' footnote corrections (R1-4, R1-5,
  R1-19 in `red/findings.md`) — every correction this corpus records **was** produced by fetching an
  additional source distinct from the one originally cited (e.g., R1-4's fix traced the 19% figure to
  a *different* paper via independent search) — so red-auditor instances, in observed practice, do
  routinely go beyond the single cited URL when a figure looks miscited. That is real, positive
  evidence the practiced behavior is better than the letter of the protocol for citation-*accuracy*
  checks. It does not, however, cover the poisoning scenario G1 is about: those corrections were
  triggered by a citation *looking wrong* (a suspicious figure), not by a citation that reads
  perfectly consistent because the same poisoned page is what both the claim and its "check" draw
  from — a consistent fabrication gives no such trigger to search further. G1 stands, narrowed:
  the practiced behavior is a real partial mitigant for citation *miscitation*, not for citation
  *poisoning by a self-consistent single source*, which is the scenario §5 item 8 names.
- G2: could `lanes = 3` simply not be intended to run with full lens-assignment engaged — i.e., is
  item 6's fix meant for a hypothetical higher default that the report just never states? Reread
  §3 rows 6 and 7 and §1.1's originating passage — no such qualifier exists anywhere; row 7's
  disposition text explicitly frames 3 as the floor to defend ("or explicit justified override"),
  with no cross-reference to item 6's redundancy requirement. G2 stands.
- G3: could the ledger's "unless the section changed" clause implicitly cover drift, on the theory
  that if an external source changed materially, a later red-lens pass would eventually notice via
  some other route (e.g., a disconfirming-evidence search) and flag the section as needing a rewrite,
  which would then re-trigger verification? Checked: this is a plausible eventual path but not a
  designed one — it depends on a *different* claim in the same section prompting an edit that happens
  to touch the sentence carrying the stale citation, which is exactly the kind of accidental,
  unenforced trigger the report criticizes elsewhere (R1-A pattern: policy without a mechanism). G3
  stands as an unaddressed structural gap, not a certainty of failure.
- G4: is this too close to R1-7/R1-8/item-11 to count as a distinct gap rather than restating them?
  Kept it, but graded LOW-MEDIUM (the lowest of this round's four) and explicitly framed as a new
  *instance* of an already-graded pattern rather than claiming novelty in kind — consistent with the
  stickler mandate to keep raising real instances rather than deduplicating them into silence, while
  not inflating its severity beyond what a repeat instance of an already-High-graded root cause (item
  11) warrants on its own.

## Not raised as gaps (checked, held to a high bar)

- Live-source drift on the report's own pinned `main` SHA (this round's own preflight discipline,
  practicing §3 row 10/[^PinnedRepoState]): `main` advanced from `47ae48d` to `88eb57f` between the
  report's last verification and this lens pass. Checked the diff — `ideas/backlog.md`, +1 line,
  docs-only, does not touch `debate.js` or any code claim this report or this lens pass relies on.
  Not a gap; noted here as the discipline working as intended.
- §3 row 16b's "keeper vs. dev/smoke" routing resolution: initially suspected a policy-without-
  mechanism gap (no explicit "keeper" flag exists in `commands/research.md`/`debate.js`). Direct
  read of `debate.js` shows `bulk = model ? { model } : {}` — when `model` is omitted (the keeper-run
  invocation the row already recommends), bulk seats (including red-lens passes) inherit the full
  session model with no override, identical to the judgment seats. The "cheapening" only fires when
  a `model` flag is explicitly passed (dev/smoke runs), which is exactly the scope row 16b's
  disposition (b) already limits it to. Checked and refuted — not a gap.
- §3 row 13's vetting-step addition (R1-14): confirmed no `.mcp.json` or CI mechanism exists to
  enforce "pin version, review source" — but the row is an unbuilt "before run 4" action item like
  most of §3, not a claim that vetting is already enforced; grading it as a fresh policy-without-
  mechanism gap would hold the report to a standard none of its other unbuilt action items are held
  to. Not raised.
