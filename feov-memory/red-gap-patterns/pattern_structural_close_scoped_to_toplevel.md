---
name: pattern-structural-close-scoped-to-toplevel
description: A structural "closed at the tool boundary" claim covers the top-level session but not the subprocess-spawned seat agents the design's own text concedes are the reachable surface
metadata:
  classes: [partial-control-coverage, cross-section-contradiction]
  type: feedback
---

A structural close (e.g. a bare `Bash` tool deny that "removes the tool entirely") is
justified on the TOP-LEVEL session's behavior ("§2.2's steps never invoke Bash"), then a
totalizing safety claim is written elsewhere ("the Bash read carve-out is closed
STRUCTURALLY … so read-scoping holds on the Bash channel too"). But the design spawns
SUBPROCESS actors — Workflow-tool FEOV seat agents, subagents — that do the actual work, and
the report's OWN text concedes Bash is reachable there ("Where Bash IS reachable (a rebuilt
rung, the Workflow seat agents, profile drift)"). The two statements contradict: the total
claim and the conceded exception cannot both hold.

**Why:** round-5 sleeper-service. R4-3's bare `Bash` deny closed the top-level session; §6
row 13 claimed the exfil channel closed structurally; §4.3 layer 4 (i) conceded seats are
Bash-reachable. On the seat surface the bare deny does nothing and closure reverts to the
non-exhaustive belt enumeration R4-3 itself called inadequate. The hook FENCE proven on seats
covers WRITES, not Bash READS — so the read+egress exfil (row 13's own threat) reopens for
exactly the nightly research actors.

**How to apply:** when a report claims a control "removes/closes X at the tool boundary,"
find the surfaces that AREN'T the top-level session — spawned subagents, Workflow/debate
seats, rebuilt rungs. Ask: does this control (a layer-1 permission RULE / `--settings`
profile) INHERIT to the spawned actor? "Hooks fire on seats" (layer 2) is NOT evidence that
the settings profile (layer 1: bare-tool deny, read-scoping, egress allow/deny) inherits.
The two are separate layers; leaf evidence for one does not transfer. Also: a fence that
blocks WRITES does not close a READ-then-egress exfil channel — check the read side
separately. Grep the report for its own "where X IS reachable" concessions and diff them
against its totalizing "X is closed" claims — the contradiction is usually already on the page.

**Now also (round-5 merge):** check the OPPOSITE assertion too — the same report can
simultaneously say seats ARE bound ("seat agents are full permission-engine + hook subjects",
layer table) and NOT bound ("Where Bash IS reachable (… the Workflow seat agents …)",
repair bullet). Raise the contradiction with BOTH horns priced: the BOUND horn is not free —
it silently starves the spawned workers of a capability the design's own methods demonstrably
use (here: citation seats' Bash git probes, live-probe lanes), and that starvation's friction
can be structurally unable to surface (ordinary-classed, waiting on corroboration that cannot
occur). "Costs the session nothing" arguments must be re-derived over the spawned population,
not just the top-level step list.
