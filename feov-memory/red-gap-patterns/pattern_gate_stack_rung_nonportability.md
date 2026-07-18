---
name: pattern-gate-stack-rung-nonportability
description: "Consent/gate stack designed for one deployment rung (local wrapper + out-of-repo settings) silently assumed to hold on other rungs (cloud routine = fresh clone, no local files; CI = settings must move in-repo)"
metadata:
  classes: [partial-control-coverage, false-universal]
  type: feedback
---

When a design offers a deployment *ladder* (manual → OS scheduler → desktop task → cloud →
CI) but its security/budget gates are artifacts of ONE rung (an operator-owned `--settings`
file outside the repo, a local wrapper script owning preflight/ledger/env-marker), audit
each rung for which layers actually survive. Cloud "fresh clone, no local files, fully
autonomous" rungs get NONE of the local profile; CI rungs invert the problem — enforcing
the profile requires committing it into the tree the loop's output can touch.

**Why:** Sleeper-service round 1 (L6-F7): §3.4's rung-3 caveat list covered qmd/connectors/
identity but never that §4.2's layer 1 and §5.1's ledger wrapper simply don't exist in a
cloud clone — an operator climbing the documented ladder carries a false belief about which
gates are on.

**How to apply:** For any ladder/matrix of deployment options, demand a per-rung
gate-survival table; treat rung upgrades as design decisions, not config toggles.

**Extension — rung ZERO is a rung (sleeper-service round 3, L5-F1):** guards keyed on the
scheduler wrapper's presence (origin-provenance stamps, ledgers, canaries) are void in the
design's own DEFAULT manual mode if manual runs bypass the wrapper — and a "manual = same
code path" claim can flatly contradict a "manual runs do not pass through the wrapper"
claim elsewhere. Audit the default/manual rung against the gate table with the same rigor
as the exotic rungs; also check the gate table itself gained rows for gates minted in
later rounds (a table built to prevent overstatement goes stale silently). Related:
[[pattern_inherited_surface_netting]], [[pattern_policy_without_mechanism]],
[[pattern_origin_tag_naming_keyed]].
