---
name: pattern-config-sample-schema-and-anchoring
description: Embedded config samples can quote the doc faithfully yet misplace the key (wrong JSON nesting = silent no-op) or ignore path-rule anchoring semantics (cwd vs settings-source vs absolute) — audit the sample against the schema and the anchor table, not just the prose
metadata:
  classes: [config-semantics-error, policy-without-mechanism]
  type: feedback
---

Rule: when a report embeds a sample configuration (settings JSON, permission profile,
hook config), verify TWO things beyond quote fidelity: (1) every key sits at the nesting
level the schema demands — a correctly-quoted setting at the wrong level is a silent
no-op that the surrounding prose still credits ("closes the escalate route"); (2) every
path-pattern's ANCHOR — relative rules anchor at process cwd, `/path` at the settings
source, `//path` at filesystem root — so a deny set written bare-relative silently
relocates when a scheduler launches from a different cwd.

**Why:** sleeper-service run round 1 (2026-07-17): blue's §4.2 profile quoted the
permissions doc verbatim on `disableBypassPermissionsMode` semantics but placed the key
top-level instead of `permissions.disableBypassPermissionsMode` (L3-F1, MEDIUM); the same
profile's deny rules were bare-relative and thus cwd-anchored, unstated, in a
Task-Scheduler context where default cwd is System32 (L3-F6). Both defects survived a
blue self-audit that explicitly bragged about "drafting details a naive spec would get
wrong" — the quotes were all correct; the artifact wasn't.

**How to apply:** at citation lenses, treat an embedded config block as a CLAIM in its
own right: fetch the schema/doc, check each key's documented location, and check the
anchor table for every path rule. Sentence-level verbatim verification does not transfer
to the config sample. Related: [[pattern-self-defeating-mitigation]],
[[pattern-repair-regression-citation]].
