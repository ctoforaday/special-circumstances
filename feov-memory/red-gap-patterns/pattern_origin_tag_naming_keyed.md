---
name: pattern-origin-tag-naming-keyed
description: an anti-self-amplification origin tag keyed on a dir-NAME/glob convention is circumvented by the process's own spawned sub-runs that land in differently-named dirs — and those sub-runs supply the "independent corroboration" the cap required
metadata:
  classes: [name-keying-vs-marker, partial-control-coverage, incomplete-repair-propagation]
  type: feedback
---

An authorship/self-poisoning guard that tags its own outputs by DIRECTORY-NAME GLOB (e.g.
`*_self-improve/`) misses artifacts the same process authors into differently-named dirs.

**Concrete instance (sleeper-service R2, L6-F2):** R1-25 capped `origin: sleeper` recurrence
at 1 and required "one non-sleeper occurrence" before a self-origin class could be picked —
tagging keyed on the `*_self-improve/` run-dir glob. But the nightly loop's step-4 spawns a
bounded FEOV smoke run in a SEPARATE dir named for the research TOPIC, not `_self-improve`.
harvest reads "every run's friction.md," so the spawned sub-run's friction (also loop-
authored) is UNTAGGED, counts at full recurrence, AND satisfies the "non-sleeper
corroboration" gate — the loop's own sub-run supplies the independent corroboration the cap
demanded from a human. The guard is circumvented by normal control flow, not an edge case.

**Why:** provenance is CAUSAL (who spawned the run), not nominal (what the dir is called). A
name-glob is a proxy that the process's own multi-dir output defeats.

**How to apply:** when blue tags self-authored artifacts to exclude/down-weight them, check
(1) does the process author into MORE THAN ONE location? (2) is the tag keyed on name/glob or
on a causal marker written at creation? Name-keyed + multi-location = hole. Worse when the
untagged sibling then UNLOCKS the capped class (supplies the required corroboration). Extends
[[pattern_authorship_evades_never_edit_guard]] (check corpus membership, not just mutation) —
here: check corpus membership by CAUSE, not by name.

**Extension — the fix closes ONE consumed surface, the amplification reappears at a SIBLING
consumed surface the tag doesn't span (sleeper-service R3, L6-F1):** R2-5 upgraded the tag
from name-glob to a wrapper-stamped `.sleeper-origin` marker on every spawned RUN DIR — sound
for run dirs. But the harvester consumes MORE than run dirs: it also mirrors red's gap-pattern
agent MEMORY and parses its headers into the same scored docket. The nightly FEOV sub-run's
red seat writes to that shared memory dir (workflow agents are memory subjects). Next harvest
mirrors the now-loop-authored patterns; they have NO run dir, hence NO marker, hence default
to NON-sleeper — re-opening the exact self-corroboration hole R2-5 just closed, one surface
over. **When a self-tag is repaired, enumerate EVERY surface the consumer reads (friction,
cost, telemetry, backlog, AND mirrored agent memory) and confirm the tag spans all of them —
a marker keyed to run-dir creation cannot tag a memory file the loop appends to.** The fix's
own scoping choice (run dirs) becomes the new blind spot's boundary.

**Extension 2 — two more origin-tag scope defects (sleeper-service R4, L5-F1/L5-F3):**
(a) **Containment-polarity inversion.** A declared out-of-contract/manual execution path's
artifacts are markerless, hence classed NON-self-origin — and the design's reassurance reads
"its output is ordinary-classed, requiring corroboration" as if that CONTAINS it. Backwards:
a non-self-origin row does not FACE the corroboration requirement, it SELF-SATISFIES it (it
IS the required independent occurrence) and corroborates tagged rows of the same class into
the pick. The untagged actor doesn't stand before the gate; it staffs it. Check the polarity
of every "requires corroboration" containment claim: who supplies the corroboration pool?
(b) **Change-keyed tagging over-strips at the wrong granularity.** Tagging any memory
file/header that "appears OR CHANGES inside a self-run window" as self-origin reclassifies a
pre-existing, human-attested pattern wholesale when the self-run merely APPENDS to it — and
append-to-existing is the memory system's own documented modal write. Result: the
best-attested classes lose their non-self standing first (monotonic docket blinding via the
guard itself; safe polarity, but signal loss). Tag what the window ADDED, not the file it
touched.
