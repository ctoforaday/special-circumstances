---
name: pattern-wildcard-allow-write-gadget
description: A wildcard allow rule characterized as "read-only" hides a write gadget via the tool's own output flag; run the leaf to find it
metadata:
  type: feedback
---

A permission allow rule with a wildcard argv (e.g. `Bash(git log *)`) described in prose as
"pinned read-only" / "no rule grants argv that chooses a write target" can still contain a
**write gadget through the tool's own output flag** — `git log --output=<arbitrary path>`
writes a file at exit 0 under the wildcard. Confirmed by running it (round 3, sleeper-service
§6 row 4 / §4.3 layer 4).

**Why:** the report enumerated only shell-level escapes (compound-command `;&&`, redirection
`>`, path traversal) in its open question (OQ18) and declared everything else "read-only." A
tool-native output flag is a fourth class that neither the enumeration nor the tamper-
watchman covers: `--output=<outside-repo>` shows in neither `git status --porcelain` nor the
guardrail-hash set, so the snapshot-compare backstop misses it.

**How to apply:** whenever blue characterizes a wildcard allow rule as read-only or argues a
risk-accept on "the argv can't choose a write target," ENUMERATE the tool's own write flags
(`--output`, `-o`, `--output-directory`, `-O`, format-to-file options) and RUN one to prove
it. A conjunctive narrowness argument ("needs a bug AND a write gadget") collapses if the
gadget is present by the rule as written, needing no bug. Check whether the design's own
tamper-evidence set (porcelain + guardrail hashes) would even see an out-of-repo target.
Related: [[pattern_per_instance_cap_not_per_cause]] (allow/deny surface reasoning),
invariant-soundness-by-enumeration (denylist under-inclusion — this is its allowlist mirror:
an allowlist rule that is over-inclusive of write channels).
