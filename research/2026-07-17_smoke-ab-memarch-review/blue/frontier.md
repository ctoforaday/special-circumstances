# blue frontier — Review the recommendations of the memory-architecture report against what has shipped since: implemented, superseded, or still open?

## H1 — Deferred as Phase 2+

Most of the blocking and High-priority items (poisoning mitigation, consolidator fixes, clone-time injection defense, provenance taxonomy, Auto Dream scope resolution) remain unimplemented; they have been deferred explicitly as a Phase 2 or later initiative, with Phase 0/1 focused on FEOV debate infrastructure, hooks inventory, and the qmd recall layer (which supersedes item 6 — the SQLite/embedding index ceiling).

## H2 — Infrastructure built, domain logic pending

Items 2, 5, 13 (agent-memory row fix, hooks test matrix, projection health) have shipped in working form; items 1, 3, 15, 16 (poisoning gates, commit-time secret scanner, clone-time injection, bootstrap down-tiering) remain open, likely staged for a follow-on PR after the infrastructure is proven in the field.

## H3 — Partially superseded by FEOV/qmd convergence

The §8 recommendations have been reformulated rather than dropped: the poisoning mitigation (item 1) is now framed as part of the "trust tier" design in FEOV debate output; the SQLite/index ceiling (item 6) is wholly replaced by qmd adoption; consolidation (items 8–11) is now addressed via RecMem + efficiency-phase patterns; hook testing (item 5) completed; Auto Dream conflict (item 4) resolved by Phase 0 being native-only (no Auto Dream consumption in v1).

## H4 — All blockers remain open; no implementation started

Every item in the blocking set (1, 2, 3, 15, 16) plus High-priority items (4, 5, 6, 7, 8, 9, 17, 20) are unchanged from the Round 4 state in the report — not started, not scheduled, and not explicitly deferred; the memory-architecture work is in a pending queue, waiting for the FEOV debate and port plan to stabilize first.

## H5 — Key items implemented, others closed by design choice

Items 2 (agent-memory row), 4 (Auto Dream scope), 5 (hook test matrix), 13 (projection health check), and 14 (OKF reframe) have been worked or resolved; the poisoning mitigations (items 1, 3, 15, 16) have been downgraded to Phase 2+ (risk-accepted for Phase 0, provided the /ingest gate is stubbed), and item 6 (index ceiling) is wholly gone (qmd adopted instead of in-process monitoring).
