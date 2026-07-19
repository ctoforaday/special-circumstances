# capture-audit — mechanized post-hoc checks (capture-research-run.mjs)

Presence/consistency tier only: these checks catch a missing line and a self-inconsistent
self-report; a plausible-but-wrong value is vacuity, whose auditor is the next run/
retrospective over these same git-tracked artifacts.

- **telemetry: FAIL** — 3 telemetry round(s) vs 13 red round(s) in debate.md
- **shards: FAIL** — measured (heuristic) closure-index lines=21, archive records=16; self-reported 15/15
- **friction-parity: REPAIRED** — 11 envelope entries auto-harvested into friction.md (labeled); 13 total now present
- **context-use: PASS** — peak 100k = 50% of its 200k window (agent a3ab165d52d3d5066); 25 seats measured; 0 over the 50% tripwire
- **assembly-screen: PASS** — 0 REFUTED-row token(s) screened against assembly-owned text; no hits
- **record-parity: FAIL** — 13 red round(s) vs 6 "### BLUE" block(s) and 4 CHANGELOG round entries (floor: redRounds-1 — a PASS exit has no final blue response)
- **record-join: FAIL** — 247 event(s) across shards; 71 distinct (seat, verb) invocations in transcripts
  FLAGGED 163 event(s) with no matching transcript invocation:
    - assemble friction (assemble:friction:#1)
    - assemble certify (assemble:certify:#1)
    - blue-lane-1 revision (blue-lane-1:revision)
    - blue-lane-1 revision (blue-lane-1:revision)
    - blue-respond-r2 position (blue-respond-r2:position)
    - blue-synthesize avenue (blue-synthesize:avenue:#1)
    - blue-synthesize avenue (blue-synthesize:avenue:#2)
    - blue-synthesize avenue (blue-synthesize:avenue:#3)
    - blue-synthesize avenue (blue-synthesize:avenue:#4)
    - blue-synthesize avenue (blue-synthesize:avenue:#5)
- **attestation-integrity: SKIP** — no anchored closures (all carried, or anchors absent)
- **record-parity-r25: FAIL** — record-parity: FAIL — 24 divergence(s):
    - ledger: open R1-7 missing from shadow — hand=open shadow=absent
    - ledger: open R1-2 missing from shadow — hand=open shadow=absent
    - ledger: open R1-9 missing from hand — hand=absent shadow=open
    - ledger: open R2-3 missing from hand — hand=absent shadow=open
    - ledger: open R2-4 missing from hand — hand=absent shadow=open
    - ledger: open R2-6 missing from hand — hand=absent shadow=open
    - ledger: closure R1-1 missing from hand — hand=absent shadow=closed
    - ledger: closure R1-2 missing from hand — hand=absent shadow=closed
    - ledger: closure R1-3 missing from hand — hand=absent shadow=closed
    - ledger: closure R1-4 missing from hand — hand=absent shadow=closed
    - ledger: closure R1-5 missing from hand — hand=absent shadow=closed

- tarball: 25 transcript(s)
- cost.md: written (telemetry join included)
- scorecards: 16 row(s) across 3 chair(s) -> feov-memory/
- precedent harvest: 9 ruling(s) -> C:\Users\gbloc\Projects\special-circumstances\law\proposed\2026-07-18_gray-area-telemetry.md (PERSUASIVE, awaiting review)
- run-live marker: removed
