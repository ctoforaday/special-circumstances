# capture-audit — mechanized post-hoc checks (capture-research-run.mjs)

Presence/consistency tier only: these checks catch a missing line and a self-inconsistent
self-report; a plausible-but-wrong value is vacuity, whose auditor is the next run/
retrospective over these same git-tracked artifacts.

- **telemetry: PASS** — 1 telemetry round(s) vs 1 red round(s) in debate.md
- **shards: PASS** — measured (heuristic) closure-index lines=3, archive records=3; self-reported 3/3
- **friction-parity: FAIL** — 1 envelope entry missing from friction.md:
    - Write-block on report.md filename triggers regardless of path; workaround = write to neutral name then cp to destination

- tarball: 9 transcript(s)
- cost.md: written (telemetry join included)
- run-live marker: removed
