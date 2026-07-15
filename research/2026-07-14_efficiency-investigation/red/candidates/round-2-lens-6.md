# Round 2 — Lens 6 (dark-side and risk: failure modes, likelihood × impact × complexity grading, security and tradeoff blindspots)

Full living report re-read in context (1178 lines, three Read windows — 25k-token cap).
CHANGELOG used as navigation hint only. All code claims below re-verified first-hand at
`plugins/frank-exchange-of-views/skills/research-protocol/scripts/debate.js` (pin-equal working
tree): PASS break l.236; contested filter ll.244–245; ONE judge dispatch per round covering the
whole docket ll.247–250; `carried` never enters `adjudicated` ll.252–253; judge full-read
prompt l.249; no `fs` import or filesystem call anywhere in the debate loop — all state rides
envelopes.

**Lens verdict input: no verdict-flipping defect. All five lever dispositions survive this
pass.** Every finding below is a failure mode inside a ROUND-1 REPAIR — the new mechanisms blue
added to close R1-1/R1-2/R1-6/R1-9/R1-12 have their own dark sides, and several share one root.

## Systemic observation (root invariant, offered for the merge's framing)

L6-F1, L6-F2, L6-F3, and L6-F5 share one root the report never states: **the engine has no
attestation primitive stronger than self-report for work-done claims.** A schema check
(required field, non-empty, valid ids) catches *omission*; a structural cross-reference between
two independent envelope structures (the shipped lineage throw: `closures` vs `gaps[].supersedes`)
catches *inconsistency*; nothing in-script can catch *vacuity* — a seat asserting work it did
not do. Round-1 repairs repeatedly reached for "make it a required envelope field" as if that
closed the class. The honest ceiling: shape checks + cross-referenced structures in-run;
independent seats and git-tracked artifacts post-run. Recommend the report state this invariant
once (one paragraph, likely §6.2 or §4.5 preamble) and name the post-hoc audit layer
(retrospective / next-run docket over git-tracked files) as the actual enforcement tier for
work-claims — a stated-invariant recommendation, not a block; severity across instances is
low-to-medium and these are hardening defects in insurance mechanisms, not verdict threats.

## Findings

### L6-F1 — condition 7's reconciliation "throw" has no executor — policy-without-mechanism, the exact class R1-10 fixed one section earlier
**Location:** §4.5 condition 7 — *"the closure index's line count is reconciled against the
archive's block count each round (one arithmetic check; a mismatch throws)."*
**Problem:** who throws? `debate.js` has no filesystem access by design (verified first-hand:
no fs usage in the loop; the report itself states the constraint twice — §3.3 "no filesystem
access by design (all state rides envelopes)", §2.5's R1-10 correction). The script cannot
count ledger lines or archive blocks. The only in-run alternative is both counts riding
`RED_ENVELOPE` — but then the same seat self-reports both sides, and the "reconciliation"
detects only a self-inconsistent self-report, not a wrong one. The round-1 repair of R1-6
commits the same policy-without-mechanism class that the round-1 repair of R1-10 fixed in §2.5
("emit into cost.md" — impossible, no fs) — same round, same author, adjacent sections.
**Grade:** MEDIUM — medium likelihood (the spec is ratified and would be built as written; the
class demonstrably recurred within one audited round) × medium impact (the observability
condition that justifies ratifying 4a cannot enforce what it claims; false assurance on exactly
the silent-failure class condition 7 exists to kill) × low complexity. Corroboration: HIGH
(code and report text side by side at this seat).
**Required fix:** name an executor that exists: (a) both counts ride the envelope AND the
reconciliation is stated honestly as a self-consistency check, with the real integrity check
named as post-hoc (retrospective/next-run red greps the git-tracked ledger and archive); or
(b) route the arithmetic to a hook — the `sc-recall-index` hook class HAS filesystem access and
already fires on every markdown write; one hook-side count comparison is the mechanism the
sentence pretends to have.

### L6-F2 — "fails structurally, not silently" overclaims: non-empty kills the unset case, not the vacuous case — R5-5's own lesson half-inherited
**Location:** §4.5 condition 7 — *"an `archive_spot_checks` field required non-empty from
round 2 (mirroring the shipped lineage throw — a merge that ran no spot-check fails
structurally, not silently)."*
**Problem:** the report's own R5-5 quote ([^R5FiveDetail]) reads "an **unset or vacuous**
supersedes leaves contested.length at 0" — two failure modes. Non-empty enforcement kills
unset; a merge that lists plausible closed-gap ids without re-verifying anything passes the
check and the run proceeds under a compliance signal. The lineage throw this clause claims to
mirror is structurally stronger: it cross-references two independent envelope structures
(`closures` against `gaps[].supersedes`, debate.js ll.231–234, verified). `archive_spot_checks`
has no independent counterpart structure in the script's sight — nothing to reconcile against.
Condition 7 converts silent failure into shape-checked self-attestation; that is an
improvement, but the sentence claims a stronger property than the mechanism has.
**Grade:** LOW-MEDIUM — medium likelihood (vacuous compliance under load is this corpus's
documented failure class — three rounds unnoticed for R3-2) × medium impact (the R5-1
stale-lineage class returns silently while the report asserts it structurally cannot) ×
trivial complexity (wording + one named post-hoc auditor). Corroboration: HIGH.
**Required fix:** reword to what the check actually guarantees ("an unset or empty field fails
structurally; vacuous entries are auditable post-hoc against the git-tracked archive — the
audit consumer is the retrospective/next-run docket") — and fold into the L6-F1/root-invariant
paragraph if adopted.

### L6-F3 — clause (v) is detective post-run, preventive never: accepted grade deltas act on spend immediately under exactly the actuation regime the interlock exists for
**Location:** §3.3 clause (v) — *"every accepted-dispute grade delta is logged in the §1.5
board-profile `log()` line... and is **spot-check-eligible** under the §4.5-condition-5 floor.
The judge is the absorber of contested deltas in both directions."*
**Problem:** the judge absorbs *contested* deltas only — an ACCEPTED dispute bypasses the judge
by construction (accepted = never docketed). Under actuation, mass is computed per merge and
sets the next round's lens budget: an accepted deflation buys blue a thinner audit **that same
round-boundary**, while the two named checks act later or are conflicted — the log line's
consumers are cost-audit.mjs and the retrospective (post-run), and the condition-5 spot-check
floor is executed by red-merge, the seat that just accepted the dispute because "accepting is
cheaper than arguing" (the report's own documented behavior). Detection arrives after the
deflation has already actuated. The interlock qualification ("valid only with (v)–(vii)")
inherits this timing hole.
**Grade:** LOW-MEDIUM — low likelihood this run (actuation rejected; zero dispute traffic
observed) rising to medium-high under the actuation the clause is mandatory for × medium-high
impact then (a full round of thinned audit before any detector fires) × trivial complexity.
Corroboration: HIGH on control flow (code + §3.3 text); MEDIUM on the incentive projection
(structural, no corpus instance — same evidentiary tier the merge accepted for R1-2).
**Required fix:** one sentence in §3.3/§3.5: under actuation, accepted deltas enter the mass
computation only after a one-round contest window, OR accepted deltas above a stated magnitude
auto-docket for judge ratification before actuating. Without it, the accepted branch is logged
but still live as a same-round deflation lever.

### L6-F4 — sharding hands the judge a filename edit and nothing else: the seat whose docket is lineage-heavy by construction loses ancestor context, and the interaction is unpriced
**Location:** §4.5 condition 6 — *"`debate.js` hardcodes `red/findings.md` at both the
red-merge prompt (l.216) and the judge prompt (l.249) — renaming without both edits strands the
judge's full read"* — and §6.1 — *"each dispatch reads `debate.md` + `red/findings.md` IN FULL
at the judgment tier (l.249): the same read profile as a merge."*
**Problem:** verified at l.249: the judge is the SECOND full-read consumer of findings.md, and
its docket is by construction the lineage-dense subset (contested = re-raised ids +
supersedes-descendants, ll.244–245) — chain rulings need the closed ancestors' full records.
Under 4a those records move to the archive. Condition 2's demanded-read MUST ("any
lineage/closure claim is verified against the archive by targeted read") is written into the
**red-auditor contract** only; no condition extends it to the lead-judge prompt or agent. A
post-sharding judge reading only the open ledger rules on supersedes chains with the ancestors'
prose absent from its read surface — and the ruling class most sensitive to missing ancestor
context is exactly `carried` vs `risk_accepted`, the gate-erosion drift path blue itself graded
MEDIUM in §6.4 item 6. Pricing is also unreconciled in both directions: §6.1's ~$10–13 judge
line assumes the unsharded full-read profile (sharding shrinks it — unpriced benefit), while
judge archive demanded-reads add cost back (unpriced cost); §4.2's savings sizing counts the
merge seat only.
**Grade:** MEDIUM — high likelihood once 4a ships (dispatch near-certain from round 2 per
blue's own §6.1; chains are the corpus norm) × medium impact (adjudication quality on exactly
the contested subset; compounds the §6.4-item-6 drift path) × trivial-to-low complexity.
Corroboration: HIGH (both prompts and the contested filter read at source this pass).
**Required fix:** (a) extend condition 2's archive demanded-read MUST to the judge (one clause
in the judge prompt / lead-judge agent: "a ruling on a supersedes-descended gap MUST read the
named ancestors' archive records"); (b) one sentence in §6.1 noting the judge line's read
profile changes under 4a, direction stated both ways.

### L6-F5 — found_by's named audit is the conflicted seat sampling itself; the independent audit exists but has no named consumer at the decision point
**Location:** §2.5 item 2 — *"`found_by` values MUST be auditable against the preserved
per-lens candidate files (grep-cheap; the files are git-tracked), and the §4.5-condition-5
spot-check floor samples them."*
**Problem:** the condition-5 floor is executed by red-merge — the seat that self-reports
`found_by` and whose lens budget a future capture-recapture throttle would set from it (R1-9's
own incentive analysis, accepted by blue round 1). The repair therefore routes the in-run audit
of red-merge's self-report to red-merge. "Auditable" (dispositional) is doing the work
"audited" (actual) should do: the grep-cheap independent audit against git-tracked candidates
has no named consumer or trigger at the one point it matters — the runs-4/5 actuation decision
the field's data feeds.
**Grade:** LOW-MEDIUM — medium likelihood (overlap attribution is clustering judgment under
incentive, not arithmetic) × medium impact (the actuation evidence base is generated by an
instrument whose only in-run check is self-sampling — the report's own
evidence-before-mechanism standard) × trivial complexity. Corroboration: HIGH
(report-internal; condition-5's executor is unambiguous in §4.5).
**Required fix:** one sentence in §2.5 item 2: any future actuation review MUST re-derive
`found_by` for a sample of gaps independently from the preserved lens files (a seat other than
red-merge — lead, retrospective, or blue), and the actuation case must cite that re-derivation.

### L6-F6 — §6.4 item 6 prices the worst case as the general case, and half its proposed fix has no mechanism
**Location:** §6.4 item 6 — *"(i) repeat judge spend, ~$10–13 per re-docketed round on the same
gap"* and *"complexity of the fix low: carry the ruling forward without re-dispatch unless
red's grade or evidence changed."*
**Problem:** (a) verified at ll.247–250: the judge dispatch is ONE agent call per round
covering the whole contested docket. A re-docketed carried gap costs a full ~$10–13 only when
it is the docket's SOLE member; whenever other gaps docket that round (the corpus norm blue
itself argues in §6.1 — dispatch near-certain from round 2), the carried gap's marginal cost is
docket-size growth inside an already-occurring dispatch, not a fresh dispatch. The efficiency
report's only new engine-defect grading mis-prices marginal cost as full dispatch cost. The
error is conservative in direction (inflates a flagged defect's cost) but this corpus corrects
conservative errors too (R1-4 precedent). (b) The proposed fix's trigger "unless red's grade or
evidence changed" is half-executable: grade changes are visible to the script in `redEnv`;
"evidence changed" lives in findings prose the script cannot read (no fs) — that half of the
trigger is policy without mechanism, and if detection falls back to red self-report, red can
force re-dispatch every round. Note also: the same carried-loop applies to clause (v)'s new
dispute-resolution enum value — a dispute the judge rules `carried` re-dockets identically; the
fix should name both traffic classes.
**Grade:** LOW — certain (textual, recomputable from the verified dispatch granularity) × low
(a grading rationale and a successor-flag spec; direction conservative) × trivial.
Corroboration: HIGH.
**Required fix:** reword (i) to "marginal docket growth inside the round's dispatch; a full
~$10–13 only when the carried gap is the sole docket member"; restate the fix trigger as
grade-change-only (mechanizable) with evidence-change handled by red re-raising with a
successor id (the existing lineage path).

### L6-F7 — clause (vi) closes one of three loop exits: maxRounds expiry is named in its own problem statement and left open; judge deadlock is a third
**Location:** §3.3 clause (vi) — *"a PASS or maxRounds expiry under the round-0 design would
moot pending disputes... Clause: dispute disposition is ordered BEFORE the PASS break, or
equivalently a PASS with pending disputes auto-dockets them for judge disposition before the
loop exits."*
**Problem:** the clause's fix sentence covers the PASS exit only. Verified exits at source:
PASS break (l.236), judge deadlock break (l.256), and `while (round < maxRounds)` expiry
(l.192). Blue speaks last in every round (blue-respond l.262 is the loop's final act), so a
dispute raised in the final round's `grade_disputes` structurally has no next red round —
maxRounds expiry moots it regardless of where disposition sits relative to the PASS break; the
deadlock break exits mid-round with any held (rejected-once, one-round-hold) disputes
undisposed. The channel's terminal case — final-round grades never adversarially audited (lane
1's §3.4, the very observation clause (vi) cites) — remains open on two of three exits.
**Grade:** LOW-MEDIUM — low likelihood (channel unbuilt; zero observed dispute traffic) ×
medium impact under the channel's own rationale (the permanent record ships with exactly the
contested grade the channel exists to protect, on the exits a strategic actor would time for) ×
trivial complexity. Corroboration: HIGH (all three exits read at source this pass).
**Required fix:** restate the clause exit-agnostically: pending or held disputes at ANY loop
exit (PASS, deadlock, maxRounds) auto-docket for judge disposition before assembly — one
sentence.

### L6-F8 — condition 6's name preflight may not exercise the guard it tests: the write-block lives in the seat harness, not the script
**Location:** §4.5 condition 6 — *"the proposed shard names MUST be test-written in the
simulator before the first sharded run, not discovered mid-merge."*
**Problem:** every observed guard firing (run-3 friction #4's scratchpad refusal; this run's
round-0 `blue/report.md` refusal at the synthesis seat) occurred at a LIVE agent seat — the
guard is a harness/permission-layer feature keyed on filename, path-independent, and its key
set is unenumerated (the report's own words). Whether "the simulator" (PR #14) spawns real
seats issuing real Write calls in the same harness, or stubs agents, is not verifiable from
this seat; if it stubs, the preflight passes vacuously and the first sharded run still
discovers the block mid-merge — the exact failure the condition exists to prevent, now behind
a green checkmark.
**Grade:** LOW — low-medium likelihood × low impact (known detour; cost is turns, not data) ×
trivial complexity. Corroboration: MEDIUM (guard behavior HIGH from two live instances;
simulator internals unverified — stated as the uncertainty it is).
**Required fix:** one clause: the preflight MUST issue real Write calls from a live seat in the
production harness (e.g., the skeleton-creation step of the next run writes both shard files),
or condition 6 must verify the simulator's writes are real before trusting a passing preflight.

## What this lens checked and did not raise

- §1.2's "no threshold setting reproduces the backlog's round-3 stop": sound within the
  severity × fix-cost parameter space the sentence states (round-2 and round-3 boards are
  indistinguishable under any severity+complexity predicate — both all-MEDIUM-HIGH-max,
  all-complexity-low). A count-keyed or board-size-keyed threshold could be tuned post-hoc to
  fire at round 3, but that is a different lever family, overfit to the target; interesting,
  not of interest. Not raised.
- §3.3 clause (vii)'s batch-docket as a blue-side judge-dispatch lever: overflow rides a
  dispatch that is near-certain anyway from round 2 (blue's own §6.1), so marginal cost ≈ 0.
  Not raised.
- Clause (v)'s log-line executability (old→new grade): reconstructable in-script from blue's
  `grade_disputes.proposed` joined with red's `dispute_responses.response` plus the prior
  round's redEnv grades — executable as specified. Not raised.
- §6.1's judge-line arithmetic and the ~$78 counterfactual recompute: consistent with cost.md
  figures as pinned. Not raised.

## Friction

- The 25k-token Read cap forced three windowed reads of `blue/report.md` (1178 lines) — the
  run-3 friction-#15 class, recurring at this seat, worsening as the living report grows.
- Simulator internals (PR #14) not inspectable from this seat within budget — L6-F8's
  corroboration is capped at MEDIUM for a claim one file-read might settle; a pointer to the
  simulator's entry point in the run skeleton would have closed it.
