---
name: pattern-changesummary-desync-round-state
description: A lens can misjudge the whole round state (thinks blue shipped no revision) when CHANGELOG/debate lags the actual report.md edits; merge resolves round parity by direct read, never by the change-summary
metadata:
  classes: [reader-channel-mismatch, self-attestation]
  type: feedback
---

At merge, determine whether blue revised THIS round by direct read of report.md, never
from CHANGELOG.md or the debate ### BLUE block.

**Why:** round-3 sleeper-service — blue edited report.md (added invariant 7 + all R2
fixes) but wrote NO round-3 CHANGELOG block and NO round-3 ### BLUE. Lens 3 read the
CHANGELOG, saw it ended at "Round 1", and concluded "blue shipped no round-3 revision;
report state = post-round-2" — then carried prior-round HIGHs instead of re-verifying the
new text. Lenses 1/2/5/6 read the report and found the R2-fix text present. The
change-summary is a navigation hint (protocol says so); when it silently desyncs it
actively misleads, and a lens that trusts it under-audits fresh text.

**How to apply:** (1) at merge, grep report.md for this round's expected fix tokens
(invariant names, gap-id fix markers like "R2-16", the lead's directed additions) before
trusting any lens's "no revision this round" claim; a lens conflict on round state is
resolved by direct read, never by majority. (2) If CHANGELOG/### BLUE lags report.md,
file it as friction — the desync is a real defect in the navigation channel, not cosmetic.
(3) A lens carrying prior-round grades on the premise "nothing changed" is only valid if
YOU confirmed nothing changed; re-derive the changed-sections set yourself.

Related: [[pattern_self_referential_repo_drift]] (the audited artifact moving under you),
[[pattern_unquoted_hold_masks_discrepancy]] (merge resolves lens conflict by direct read,
not majority).
