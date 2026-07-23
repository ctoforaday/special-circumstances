# capture-audit — mechanized post-hoc checks (capture-research-run.mjs)

Presence/consistency tier only: these checks catch a missing line and a self-inconsistent
self-report; a plausible-but-wrong value is vacuity, whose auditor is the next run/
retrospective over these same git-tracked artifacts.

- **telemetry: FAIL** — board-telemetry.jsonl absent; debate.md shows 2 red round(s)
- **shards: SKIP** — no ledger/archive (pre-sharding run)
- **friction-parity: REPAIRED** — 3 envelope entries auto-harvested into friction.md (labeled); 3 total now present
- **context-use: PASS** — peak 98k = 49% of its 200k window (agent ae430a4743f89fb85); 8 seats measured; 0 over the 50% tripwire
- **assembly-screen: PASS** — 0 REFUTED-row token(s) screened against assembly-owned text; no hits
- **record-parity: PASS** — 2 red round(s) vs 1 "### BLUE" block(s) and 1 CHANGELOG round entry (floor: redRounds-1 — a PASS exit has no final blue response)
- **record-join: FAIL** — 50 event(s) across shards; 17 distinct (seat, verb) invocations in transcripts
  FLAGGED 4 event(s) with no matching transcript invocation:
    - blue-lane-1 avenue (blue-lane-1:avenue:#1)
    - blue-lane-1 avenue (blue-lane-1:avenue:#2)
    - blue-lane-1 avenue (blue-lane-1:avenue:#3)
    - blue-lane-1 avenue (blue-lane-1:avenue:#4)
- **attestation-integrity: SKIP** — no archive records to reconcile
- **record-parity-r25: FAIL** — record-parity: FAIL — 3 divergence(s):
    - debate: ### RED section count — hand=2 shadow=0
    - debate: ### BLUE section count — hand=1 shadow=0
    - citation-ledger: row count — hand=9 shadow=0

- tarball: 8 transcript(s)
- cost.md: written (telemetry join included)
- scorecards: 16 row(s) across 3 chair(s) -> feov-memory/
- precedent harvest: no rulings this run
- run-live marker: removed
