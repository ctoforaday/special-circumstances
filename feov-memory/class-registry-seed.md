# Class registry seed (E0.5g, 2026-07-18) — 224 gaps, 3 runs, 38 classes

Full analysis: the E0.5g report (this file is the durable seed W2d consumes).
Corpus: retrospective 52 + efficiency 77 + sleeper 95 gaps. Caveat: efficiency
rounds 1-3 boards were destroyed by in-place overwrite (only fed2449 committed);
~20 assignments reconstructed from lens-candidate headers (±2-3 per class) — a
live instance of the artifact-preservation class this registry itself tracks.

## Headline numbers

- 218/224 gaps (97%) land in a ≥2-instance class; 93% in ≥3-instance classes;
  only 6/38 classes are true singletons. THIS INVERTS the design assumption
  ("most classes are singletons") — see design consequences below.
- Top-5 classes appear in 3 of 3 runs (topic-independent generators):
  citation-figure-misattribution (20), figure-recount-fails (15 — 5/5/5 across
  runs), incomplete-repair-propagation (15), policy-without-mechanism (15),
  derivation-status-overclaim (11).
- recurrence-detector-keying — the class that indicts the escalator's own keying
  — is itself 3-for-3 across runs.
- NEW classes red's memory lacks: name-keying-vs-marker (3, all sleeper),
  config-semantics-error (2), cross-corpus-id-collision (1), negative-definition
  (1 primary — but root of the 11-gap sleeper provenance chain; severity-weight
  far exceeds count).

## The 27 multi-instance classes (slug — count FEOV/EFF/SLP)

citation-figure-misattribution 20 (4/10/6) · figure-recount-fails 15 (5/5/5) ·
incomplete-repair-propagation 15 (4/5/6) · policy-without-mechanism 15 (4/7/4) ·
derivation-status-overclaim 11 (1/6/4) · cross-section-contradiction 10 (3/1/6) ·
enumeration-non-exhaustive 9 (0/0/9) · unhandled-degenerate-case 9 (2/1/6) ·
false-universal 8 (4/1/3) · unverified-composition 8 (1/3/4) · self-attestation 8
(0/2/6) · spec-underspecification 8 (0/4/4) · claim-contradicts-own-record 7
(1/4/2) · reader-channel-mismatch 7 (1/3/3) · exhaustive-sweep-omits-case 7
(0/4/3) · risk-coverage-omission 7 (3/0/4) · live-source-drift 6 (4/0/2) ·
within-source-condition-misattribution 5 (2/3/0) · partial-control-coverage 5
(0/3/2) · figure-miscomposition 4 · recurrence-detector-keying 4 ·
pricing-basis-drift 4 · metric-conflation 4 · citation-status-drift 3 ·
undecided-disjunction 3 · causal-narrative-fails-reproduction 3 ·
name-keying-vs-marker 3. Two-instance: reflexivity-blindspot,
missing-root-invariant, audited-artifact-sibling-halo, artifact-preservation,
config-semantics-error. Singletons: vote-laundering,
verification-scope-blindspot, doctrine-vs-implementation,
measurement-methodology-drift, cross-corpus-id-collision, negative-definition.

## Design consequences for W2d (bind these before building)

1. FAT HEAD, NOT SINGLETON TAIL: with top classes at 15-20 instances/corpus, an
   N=2 recurrence escalator would docket constantly. Recalibrate: recurrence
   counts WITHIN-RUN after the class last reached zero open instances (the
   original design's intent), and the escalator's generator-fix demand keys on
   within-run re-spawning, not corpus totals. Cross-run class pressure routes to
   craft memory (manifest lines / lens duties), not the docket.
2. LINEAGE IS ORTHOGONAL TO CLASS: a successor inherits the class of its
   PROXIMATE defect, not its chain — supersedes carries lineage; class carries
   kind. The anti-shell-game inheritance rule applies to the chain dimension.
3. TIE-BREAK RULES REQUIRED at minting time for adjacent classes
   (false-universal vs enumeration-non-exhaustive vs exhaustive-sweep-omits-case;
   cross-section-contradiction vs claim-contradicts-own-record) — the registry
   entry for each names its neighbor and the distinguishing question.
4. The registry seed doubles as blue's manifest-line source and red's lens-duty
   source: top classes are exactly what the correctness manifest's rows check.
