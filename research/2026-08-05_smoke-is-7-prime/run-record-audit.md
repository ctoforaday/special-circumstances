# capture-audit — mechanized post-hoc checks (capture-research-run.mjs)

Presence/consistency tier only: these checks catch a missing line and a self-inconsistent
self-report; a plausible-but-wrong value is vacuity, whose auditor is the next run/
retrospective over these same git-tracked artifacts.

- **telemetry: PASS** — 2 telemetry round(s) vs 2 red round(s) on the record
- **shards: SKIP** — no ledger/archive (pre-sharding run)
- **friction-parity: FAIL** — 3 envelope friction entries never reached the record (should have been recorded via the friction verb during the run):
    - KNOWN HARNESS LIMIT W1.11 (re-logged for tracking): Glob/Grep refuse paths outside registered working directories while 
    - Tool issue: feov-record blue manifest-row --id R1-2 rejected as 'dangling reference' (gap not minted by blue in current 
    - PostToolUse hook reports removal of 5 citation anchors (c-10dbc13b, c-043f47d1, c-2e414f45, c-b7e8192a, c-13a3be75) that
- **context-use: WARN** — peak 102k = 51% of its 200k window (agent a812061df34f0f58d); 15 seats measured; 2 over the 50% tripwire:
    - a812061df34f0f58d: 102k / 200k
    - a9fe0d7d17bf3af01: 102k / 200k
- **assembly-screen: PASS** — 0 REFUTED-row token(s) screened against assembly-owned text; no hits
- **record-parity: PASS** — 2 red round(s) vs 3 blue sitting(s) and 3 CHANGELOG round entries (floor: redRounds-1 — a PASS exit has no final blue response)
- **record-join: FAIL** — 107 event(s) across shards; 48 distinct (seat, verb) invocations in transcripts
  FLAGGED 25 event(s) with no matching transcript invocation:
    - blue-respond-r1 blue_edit (blue-respond-r1:blue_edit:#1)
    - blue-respond-r1 blue_edit (blue-respond-r1:blue_edit:#2)
    - blue-respond-r1 blue_edit (blue-respond-r1:blue_edit:#3)
    - blue-respond-r1 blue_edit (blue-respond-r1:blue_edit:#4)
    - blue-respond-r1 blue_edit (blue-respond-r1:blue_edit:#5)
    - blue-respond-r1 blue_edit (blue-respond-r1:blue_edit:#6)
    - blue-respond-r2 avenue (blue-respond-r2:avenue:#1)
    - blue-respond-r2 avenue (blue-respond-r2:avenue:#2)
    - blue-respond-r2 avenue (blue-respond-r2:avenue:#3)
    - blue-respond-r2 avenue (blue-respond-r2:avenue:#4)
- **attestation-integrity: SKIP** — no archive records to reconcile
- **model-tier: PASS** — every seat ran on its configured tier (bulk=haiku, judgment=haiku)

- tarball: 15 transcript(s)
- cost.md: written (telemetry join included)
- report.md: cost breakdown folded in (## Cost)
- scorecards: 18 row(s) across 3 chair(s) -> feov-memory/
- precedent harvest: 1 ruling(s) -> C:\Users\gbloc\Projects\special-circumstances\law\proposed\2026-08-05_smoke-is-7-prime.md (PERSUASIVE, awaiting review)
- run-live marker: removed
