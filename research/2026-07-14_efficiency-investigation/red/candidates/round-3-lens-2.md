# round 3 — lens 2 (leaf-node citation verification, slice 2: §2–§3)

Full living report re-read whole (1365 lines, three windowed reads). CHANGELOG round-2 edits
landing in this slice: R2-1 (§2.5 item 1 + [^JournalCheck]), R2-2 (§2.4), R2-6 (§3.3 v/vi),
R2-7 (§2.5 item 2), R2-12 + R2-18 (§2.2), R2-14 (§2.3 + [^Sprt]). Every repair re-verified as
a new claim (repair-regression discipline); unchanged HIGH-ledgered claims carried per the
ledger rule. Ledger appended (6 entries, round-3 lens-2 block).

## Repair verifications (all faithful — no repair-regression found this round)

1. **[^Sprt] / §2.3 (R2-14)** — re-fetched arXiv:2603.00216. Abstract verbatim: "for symmetric
   error bounds, the sequential test reduces the average sample size by at least 36\% and by
   at most 75\%." Blue's restored quotation matches in full, second "by" included, condition
   attached. **HIGH.**
2. **[^JournalCheck] / §2.5 item 1 (R2-1)** — three sub-claims leaf-verified first-hand:
   (a) debate.js `log(\`researching: ...\`)` present at the l.52 region (direct read);
   (b) this run's LIVE journal (wf_5cefd2a4-35f) re-checked at round 3: now 50 lines =
   28 `started` + 22 `result`, grep "researching" = 0 — blue's mid-round-2 snapshot (43 =
   22+21) is consistent with growth and the console-ephemeral conclusion HOLDS on the live,
   moving artifact; (c) cost-audit.mjs (actual path:
   plugins/frank-exchange-of-views/skills/research-protocol/scripts/cost-audit.mjs) input
   filter is `startsWith('agent-') && endsWith('.jsonl')` with zero journal.jsonl references.
   **HIGH** on all three. Run-3 journal 87 = 46+41 carried HIGH (R2 merge block).
3. **§2.4 (R2-2)** — "rounds 1 and 2 both demonstrably ran 6": red/candidates holds exactly
   round-{1,2}-lens-{1..6}.md (12 files, ls first-hand). Arithmetic: ceil(160/40)=4+2=6 seats;
   3 removed agents × ~$2 × 3 rounds = ~$18; ~10% of the ~$180 rescaled baseline ✓. **HIGH**
   on mechanism and arithmetic (basis note → L2-F2 below).
4. **§3.3(vi) (R2-6)** — three loop exits ll.192/236/256 carried HIGH from the round-2 merge
   block (first-hand then); clause text matches. Clause (v) contest-window residual (delta
   accepted in the final round ships un-reviewed) is lane 1's §3.4 final-round-grades
   observation, already carried in the report — no new gap.
5. **§2.2 (R2-12, R2-18)** — subject restored; "403 shows access failure, not mechanism —
   bot-block vs paywall unshown" is exactly the demanded honesty; "the journal is
   subscription" hedged-plausible (Elsevier Computers & Security). **HIGH / MEDIUM-HIGH.**
6. **§2.5 item 2 (R2-7)** — design clause; internally consistent with the §6.2 attestation
   ceiling it cites (in-run red-merge sampling labeled self-report; off-seat re-derivation
   named as the actuation-review requirement). No citation to falsify. **Consistent.**

## Findings

### L2-F1 — §3.3 default-to-docket price not repriced under the accepted R2-11 marginal-docket logic — LOW

- **Location:** §3.3 (lever 3a mechanism), default-to-docket bullet: "auto-escalates every
  open dispute to a judge dispatch at ~$10/firing (the backlog's only judge-round estimate)."
- **Challenge:** Round 2 repriced judge-docket traffic (red R2-11, accepted into §6.4 item 6):
  the dispatch is ONE agent call per round covering the whole docket, so a re-docketed item
  costs the full ~$10–13 only as the docket's sole member — otherwise marginal growth. §6.4
  item 6 even names dispute traffic as the second class the loop covers ("a dispute ruled
  carried re-dockets identically — both traffic classes named"), yet §3.3's price line still
  carries the pre-R2-11 per-firing figure. Stale-baseline pricing: a repricing insight
  accepted in one section, unpropagated to the sibling price in another.
- **Grading:** likelihood certain (textual inconsistency, side-by-side read) × impact low
  (direction is conservative — it overstates the cost of a design blue ratifies anyway; no
  disposition turns on it) × complexity trivial (one clause: state the marginal-growth pricing
  per §6.4 item 6, full ~$10 only when the docket is otherwise empty).
- **Required fix:** reprice §3.3's default-firing cost consistently with §6.4 item 6's
  marginal-docket-growth rule, cross-referencing it.

### L2-F2 — §2.4's "× 3 rounds" basis unstated in-section — LOW

- **Location:** §2.4: "cutting 4→1 citation instances removes ~3 agents/round ≈ ~$6/round at
  run-3 per-lens rates × 3 rounds = ~$18/run."
- **Challenge:** the 3-round multiplier (vs run 3's 5 rounds) is an assumption about WHICH
  rounds a mass-scaled throttle would actually throttle (presumably the low-mass rounds 3–5
  per §2.1's series), inherited silently from the round-0/round-1 construction. Arithmetic
  checks; the basis does not. A reader cannot recompute the estimate's scope from the section.
- **Grading:** likelihood certain (textual) × impact low (figure is self-graded MEDIUM, the
  REJECT rests on §2.1/§2.3, and the direction question doesn't change) × complexity trivial
  (five words naming the throttled-rounds assumption).
- **Required fix:** state the basis, e.g. "× 3 throttled rounds (the low-mass rounds 3–5 of
  §2.1's series)."

### Out-of-slice observation (for merge routing, not a slice finding)

§6.4 item 6 (slice 4 territory) refers to "§3.3 clause (v)'s dispute-resolution enum" — clause
(v) is the accepted-branch audit trail and contains no enum; the dispute resolution enum
(`accepted|rejected`) lives in §3.3's RED_ENVELOPE bullet and the judge's resolution enum in
its final bullet. Pointer defect, trivial; anchored in §6.4 so it belongs to the slice-4/merge
read.

## Slice verdict

§2–§3 round-2 repairs: all six faithful at the leaf, zero repair-regressions. Two LOW
textual/pricing-consistency findings minted; nothing graded above LOW. No friction this pass.
