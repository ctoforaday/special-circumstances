---
name: pattern-ephemeral-instrument-and-grid-count
description: two quantitative-claim hazards — a measurement whose parser/input lives only in scratchpad (self-report with no re-derivation path), and table counts asserted by grid arithmetic instead of row enumeration
metadata:
  classes: [artifact-preservation, figure-recount-fails, self-attestation]
  type: feedback
---

Two leaf-verification hazards on quantitative claims, both caught round 3 of the efficiency run.

1. **Ephemeral instrument**: blue "measures" something first-hand (good) but the measuring
   script lives only in a session scratchpad and its input is an untracked working-tree file.
   The figures become self-report with no independent re-derivation path — exactly the vacuity
   tier the system's own attestation ceiling routes to "post-hoc audit over git-tracked
   artifacts," which these artifacts aren't. **Why:** [^MergeDecomposition]'s ~70-line parser
   (never committed) + gitignored tarball fed §4.2's table, the money map's #1 ranking, and an
   ANSWERED open question. **How to apply:** when a report cites its own measurement, ask where
   the instrument and input live NOW; "method stated in prose" ≠ preserved. Grade as
   preservation gap (usually LOW-MEDIUM, trivial fix: commit the script), not corroboration
   failure. Related: [[pattern-provenance-self-report-and-stale-gate]].

2. **Grid-count fetch hazard**: verifying "N reported configurations" against a paper's table,
   a summarizing WebFetch will assert the dimensional product (6 datasets x 4 models = 24) even
   when cells are missing (multimodal rows ran only 3 vision-capable models; true count 22).
   **Why:** first fetch of arXiv:2510.12697 Table 2 said 24/24 while its own enumeration
   totaled 22 — the enumeration was right, the header wrong. **How to apply:** never accept a
   count from a fetch summary; demand per-row enumeration and sum it yourself. A blue footnote
   that says 22 against a table that "obviously" reads 24 may be the CORRECT one.
