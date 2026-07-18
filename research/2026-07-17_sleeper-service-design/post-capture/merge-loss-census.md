# E0.5b — Red-merge bottleneck census (2026-07-18)

Corpus: 68 lens findings across three sampled rounds (sleeper r2 + r4, efficiency r4)
traced into ledgers/archives/debate records.

## Verdict: the merge is essentially LOSSLESS at the finding level

- DROPPED-SILENT: 0 in all three rounds. Every L*-F* finding traces to a minted gap
  or a recorded declination. Compression (26->22, 21->16, 20->10) is entirely
  declared composites of concurring lenses. Merge-minted-without-lens-source: 0
  (merge-original content is verification and one openly-contested adjudication —
  a seventh verifier, not a seventh finder).
- Grade fidelity: 34/35 verifiable minted grades faithful to lens grades; 2
  reconciliations recorded (one softening, one hardening); 0-1 unrecorded.

## The two real findings

1. NOTE-CLASS OBSERVATIONS HAVE NO FATE TRACKING — the only true loss surface.
   Lens "note only" items exist solely in candidate files: the R3-11
   over-normalization note, the date-keyed idempotency interplay ("worth one line
   in scheduling.md" — never landed), and an L4 "immaterial" note later vindicated
   by the R4-2 fix itself. One r2 declination (FEOV 0.6.0 fragment) was the wrong
   call, deferred-caught at r3.
2. THE LOSS SURFACE MIGRATED FROM MERGE TO RECORD RETENTION: ledger.md is
   rewritten in place, so per-round board states are unrecoverable — 3 of round
   4's 16 minted grades (R4-5/6/9) can no longer be checked against their lens
   grades AT ALL (~19% of that round's grade decisions unauditable by record
   loss, not by merge behavior). The R3-14 lossy-summary class is structurally
   unfalsifiable for past rounds for the same reason.

## Record-layer requirements derived (feeds W2f)

(1) found_by at LENS-FINDING granularity ([L5-F3, L6-F2], not [L5, L6]);
(2) contributing lens grades printed beside the minted grade on the gap record;
(3) per-round ledger snapshots (append-only or committed per round, as archive.md
already is); (4) a one-line merge disposition for EVERY candidate-file observation
including notes and checked-and-held items: minted-as | folded-into | declined:
<reason> | banked. The efficiency run's single findings.md is the better template
for several of these and predates the sharding.
