---
name: pattern-closed-mode-builtin-carveout
description: A design's "default-closed / allowlist-inverted" premise built on a platform mode is defeated by the mode's own documented built-in auto-approve carve-out
metadata:
  type: feedback
---

When a report claims a permission MODE makes the surface default-closed ("anything not
allow-listed is denied"), verify the mode's documented EXEMPTIONS, not just its headline
semantics. Instance (sleeper-service r3, L2-F2): `dontAsk` "denies anything not in your
permissions.allow rules **or the read-only command set**" — a non-configurable built-in Bash
set (`ls, cat, echo, pwd, head, tail, grep, find, wc, which, diff, stat, du, cd`, read-only
git) that runs "without a permission prompt in every mode". The design's read-scoping
closure (Read/Grep/Glob allow-scoped + Read denies) covered the Read TOOL only; Bash `cat`
on the same secret paths was auto-approved, silently re-opening a risk-matrix row already
graded closed.

**Why:** closed-world premises inherit every platform carve-out; the carve-out lived on the
same doc pages a prior round had graded HIGH — either live-source drift or under-read, so
re-fetch mode semantics whenever a closed-world claim leans on them.

**How to apply:** (1) for any "default is closed" claim, grep the platform docs for the
mode's exception clauses ("except", "or the", "in every mode", "not configurable");
(2) smell: ALLOW rules for things the carve-out auto-approves anyway (redundant allows =
unmodeled carve-out); (3) check whether the carve-out's members compose with retained
egress (read-only reads + WebFetch = exfil); (4) carve-out members with operator/redirect
forms (`echo >>`) belong in the same shell-operator build-test as the design's other Bash
rules. Related: [[pattern-invariant-soundness-by-enumeration]],
[[gap-live-source-drift]].

**Extension (r4, L2-F2 — the carve-out story cuts BOTH ways):** the same doc that grants
the carve-out can grant deny-rules reach INTO it — "Read and Edit deny rules apply ... to
file commands Claude Code recognizes in Bash, such as cat, head, tail, and sed." A repair's
motivating example ("cat <transcript path> would have been AUTO-APPROVED under the old
profile") failed at the leaf because the old profile already carried an explicit Read deny
on that exact path, which the doc extends to Bash cat. The carve-out exposure is real only
for paths protected by ALLOW-SCOPING alone (no deny rule). So: when grading a
carve-out/bypass severity story, test the chosen example against the FULL rule-interaction
matrix (deny reach, deny-beats-carve-out), not the carve-out clause in isolation — the
prior-exposure narrative is a claim to reproduce, same as the repair
([[pattern-postmortem-misdiagnosis]]).
