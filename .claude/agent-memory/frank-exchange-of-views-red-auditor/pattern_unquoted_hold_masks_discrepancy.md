---
name: pattern-unquoted-hold-masks-discrepancy
description: A lens's "checked, held, not raised" entry that asserts a match WITHOUT quoting both sides can mask a live discrepancy; when lenses conflict, the merge seat decides by direct read, never by majority or by trusting the hold
metadata:
  type: feedback
---

Rule: a hold-claim ("verified: X matches Y, no discrepancy") carries evidentiary weight only if
it quotes both texts side by side, the same standard gap-claims are held to. An unquoted hold is
a weak negative and MUST NOT rebut a sibling lens's quoted positive finding.

**Why:** FEOV retrospective round 5 — lens 5's "noted, checked, not raised" item asserted §3
row 23's regression-chain enumeration "matches §2.1's text exactly; no discrepancy found," while
lenses 1, 2, and 4 each independently quoted row 23 carrying blue's discarded, factually-wrong
first-pass list. Direct read at the merge seat (report lines 496 vs. 727) confirmed the three
quoting lenses right and the hold wrong. Same round, second instance: lens 2's ledger line
asserted a five-id list "matched" a six-id set — mechanical extraction disproved it. Both errors
were inside *verification-positive* statements, the kind that never get adversarial pressure
because they raise no gap.

**How to apply:** at every red merge, (1) when any lens's hold contradicts any sibling lens's
finding, resolve by first-hand read of the primary text — never by lens majority, seniority, or
plausibility; (2) treat "verified clean" ledger lines that name a specific correspondence
(list X = list Y, text A = text B) as claims to re-extract mechanically (grep/sed the exact
line), not to trust; (3) in my own lens passes, quote both sides in every held comparison — a
hold without quotes is homework not done. Related: [[pattern-repair-regression-citation]]
(red-side errors get logged with the same discipline demanded of blue),
[[pattern-schema-legal-control-flow-trace]].

**Extension (sleeper-service r1):** the hold can come WITH a mechanism story that itself was
never verified. Lens 1 saw wc=1,557 vs report's "1,558 lines" and ruled "trailing-newline
artifact, not a discrepancy" — but `tail -c 1` shows the final byte IS a newline, so the file
has exactly 1,557 lines and the excuse fails reproduction (kin to
[[pattern-postmortem-misdiagnosis]]). An artifact-excuse for a numeric mismatch (trailing
newline, encoding, off-by-one-of-counting-method) is a testable claim: test it at the merge
seat before accepting the hold. Cost: one shell byte-check.
