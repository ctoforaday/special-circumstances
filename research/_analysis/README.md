# _analysis — one-off analysis tooling, rescued from session scratchpad

These are the scripts and drafts that produced the findings in
`research/2026-07-17_sleeper-service-design/post-capture/` and the E0.5
corpus-exhaustion series. They lived only in a session scratchpad under
`%TEMP%` and would have been lost when temp was cleaned; they are here for
reproducibility, not as shipped code. Nothing in the plugins imports them.

| File | What it did |
|---|---|
| `gap-class-proposal.md` | The E0.5g analysis over 224 gaps that distilled into `feov-memory/class-registry-seed.md` (38 classes). The seed is the product; this is the working. 97% multi-instance — the finding that inverted the singleton assumption. |
| `dash-watch.mjs` | The live run dashboard watcher (run 5). Must be launched with cwd = this repo — the run-live marker is cwd-keyed. |
| `judiciary-analysis.mjs` | Judiciary-section stats: rulings by type, disputes, chain-aware argument longevity (PR #27). |
| `compare-smokes.mjs` | Smoke-run A/B comparison (baseline vs memarch-review). |
| `indexing-correlation.mjs` | Correlation probe behind the E9 embedding-assist musing. |
| `doctor-transcripts.mjs`, `doctor-config.mjs` | Transcript/config scans for the doctor checks. |
| `apply-permissions.mjs` | Batch permission-rule application to settings. |
| `probe-timings.mjs` | Hook/probe timing measurements. |
| `run5-dashboard-artifact.html` | Source of the published run-5 dashboard artifact. |

Provenance: session `44104cac-198f-40ac-84a1-3f30486755f7`, scratchpad under
the AgentOrange project dir, moved 2026-07-18 when work relocated to this repo.
