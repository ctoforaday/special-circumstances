---
name: pattern-sibling-repair-composition
description: Two same-round repairs to one design (mechanism hardened, registered test/prediction restated) can fail to compose — re-derive the test from the hardened mechanism
metadata:
  classes: [unverified-composition, incomplete-repair-propagation]
  type: feedback
---

When one round applies two fixes to the same paragraph/design — one hardening the MECHANISM
(e.g. arm conditions upgraded to double confirmation) and one restating its REGISTERED
TEST/PREDICTION (e.g. dollar netting corrected) — verify they compose: the prediction's
trigger condition must be re-derived from the hardened mechanism, not carried over from the
pre-hardening version.

**Why:** efficiency-investigation round 3 (L5-F2): R1-12 hardened the re-scoped floor's arm
to "two consecutive zero-above-floor-mint rounds," while R1-17 restated the registered
prediction's netting but left its trigger single-round and total-mint-keyed. Both closed
clean in round 2 individually; jointly the registered test could settle TRUE while the
hardened variant saves $0 — falsely validating a build trigger. Sibling of
[[pattern-unreconciled-numeric-floors]] (requirements that don't arithmetically compose) but
at the repair layer: each repair verifies clean in isolation; the defect is only visible
reading them together.

**How to apply:** when auditing repairs, group same-section fixes and check pairwise
composition — especially mechanism-change + registered-figure/prediction pairs, and any
"arm/trigger" clause vs the test that claims to measure it. A repair-verification sweep that
checks each closure independently (as round-2 did) structurally misses this class.
