# E0.5a — Attestation verification: red's testimony vs behavior (2026-07-18)

Method: 27 archive closure records sampled (all method classes, all 5 rounds, ALL
maximal-strength claims included) + 5 citation-ledger rows, each cross-checked against
the attesting seat's ACTUAL tool calls (per-seat tool-call index extracted from all 58
run-5 transcripts). Auditing behavior, not testimony.

## Tally

- CORROBORATED: 23/27 — including every maximal-strength claim (byte-identical-at-pin,
  leaf-reproduced-on-this-box, live-fetch-zero-drift); 6 exact to the probe path
  (l5probe, l4probe_r5, l6probe, r5mergeA — red's lens-tagged temp paths are a de facto
  behavioral signature).
- PARTIAL: 1 — R2-13: merge claims "direct read," behavior shows grep-over-concatenated-
  cat; artifact right, act adjacent.
- UNCORROBORATED: 1 leg — R1-7's round-3 "re-confirmed 6df52af-clean this session
  (r3 L1)": no round-3 red seat touched the AgentOrange repo at all. The r2 and r5
  checks are real (r5's is corroborated exactly) — a CARRIED result re-attested as a
  fresh same-session act. Inflation rate ~4% of records, ~1 in 30 checkable legs.
- UNRESOLVABLE: 3 — "(CLEAN)", unanchored arithmetic "recomputed" x2.

## Per-method rates

Live fetch 8/8; pin checks (git show/cat-file) 6/6; leaf-reproduction probes 4/4 with
verbatim paths; "direct read" always letter-true but near-zero information (every seat
reads the report); "recomputed" is BIMODAL — 100% checkable when it names a command
(wc, grep recounts), 0% when pure arithmetic; "spot-check/(CLEAN)" unauditable as
phrased. Citation-ledger spot-checks: 5/5 corroborated, several exact; the ledger even
honestly discloses its own tool degradation (gh classifier-blocked at seats).

## Judgment and prescription

Red's attestation record is honest at the behavioral level. The reformed constitution
does NOT need more verification machinery — the machinery demonstrably runs — it needs
an ATTESTATION-FORMAT INVARIANT: (1) every "verified" clause carries seat + tool +
target (command or URL); (2) "recomputed" shows the recount command or the arithmetic
inline; (3) re-attestation of a prior round's result is labeled CARRIED (citing the
round whose tool act grounds it), never performed; (4) merge-relayed reads name the
reading seat. With these, the record would have been ~100% mechanically auditable
instead of ~89%. Folded into red's constitution (W2a, PR #32).
