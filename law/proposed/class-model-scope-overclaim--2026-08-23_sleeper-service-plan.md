# proposed gap class — `model-scope-overclaim` [PROPOSED — not in the registry until adopted]

Coined by `red-merge-r2` during `2026-08-23_sleeper-service-plan`. It is NOT staged into any later run:
an unreviewed class validating a future `--class` is the registry losing the only thing it
means. Adopting it means adding the slug to `feov-memory/class-registry.json` by hand.

- **definition**: A mathematical or model-based claim asserts a generality (e.g. shape-invariance across parameter regimes) broader than the underlying equation actually establishes once its parameters are varied
- **neighbour**: `derivation-status-overclaim`
- **distinguisher**: derivation-status-overclaim is about how many independent sources/lanes corroborate a finding; model-scope-overclaim is about whether a claimed mathematical property (e.g. 'the shape is robust to X') matches what the model equation actually yields when X is varied, versus conflating it with a different, unstated parameter
- **first used on**: R2-1

The three fields above are the seat's own words, refused at the write if any were missing
(`record.validateClassNew`), so this proposal is well-formed by construction. What a
reviewer decides is whether it DISCRIMINATES — whether the distinguisher actually separates
it from its neighbour on a case where both are arguable.
