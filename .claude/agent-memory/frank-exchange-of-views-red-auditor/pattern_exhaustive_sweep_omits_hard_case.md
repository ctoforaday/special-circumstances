---
name: pattern-exhaustive-sweep-omits-hard-case
description: A self-certifying completeness sweep ("every catch traced", "checked against every position", "both misattributions corrected") silently omits the hardest named case — often one the same report itself elevated pages earlier
metadata:
  type: feedback
---

When a report contains an exhaustiveness claim — a mapping table ("every late-round catch
traced to an arm"), a systematic check ("doctrine check run against every position"), or a
counted sweep ("two frontier misattributions corrected") — enumerate the claimed universe
independently and diff it against the table/check/count. The characteristic defect: the
omitted item is the *hard case*, and frequently one the report itself named as a type
specimen or conceded as a conflict in an earlier section (writer's blind spot: the case was
"handled" mentally when first discussed, so the sweep skips it).

Observed (run-4 efficiency investigation, round 1): §5.2 "every late-round catch" table
omitted R5-2 — one of §5.1's own two named specimens, and the only catch NO proposed scope
arm covers (L5-F1, MEDIUM-HIGH); §6.3 doctrine check "run against every position" skipped
the §5.5 conditional-ratify, the one position §5.3 itself admits conflicts with a named
doctrine clause (L5-F2); §6.4.3 counted "two" frontier misattributions while a third (stale
pre-temper grades in H1) survived (L5-F10).

**Why:** exhaustiveness claims are load-bearing for gate/ratification logic — a sufficiency
argument with a silent hole propagates into future-run dispositions uncritically.

**How to apply:** at the logic/completeness lens, treat every "every/all/both/N of N"
sentence as a checkable enumeration: build the universe from the report's OWN earlier
sections (specimens it names, positions it takes, errors it admits), then verify membership
one by one. The omitted member is usually adjacent to a concession. Related:
[[pattern-missing-root-invariant]], [[pattern-unreconciled-numeric-floors]].
