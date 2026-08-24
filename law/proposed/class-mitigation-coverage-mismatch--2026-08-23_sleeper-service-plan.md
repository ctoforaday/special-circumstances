# proposed gap class — `mitigation-coverage-mismatch` [PROPOSED — not in the registry until adopted]

Coined by `red-merge-r2` during `2026-08-23_sleeper-service-plan`. It is NOT staged into any later run:
an unreviewed class validating a future `--class` is the registry losing the only thing it
means. Adopting it means adding the slug to `feov-memory/class-registry.json` by hand.

- **definition**: A real, working mitigation is assigned to a failure mode, but the mitigation's temporal or scope coverage does not match the failure's actual recurrence or extent, so the failure can still land inside the mitigation's blind window
- **neighbour**: `policy-without-mechanism`
- **distinguisher**: policy-without-mechanism means the required property has NO enforcing mechanism at all; mitigation-coverage-mismatch means a mechanism genuinely exists and works, but only checks once (or narrowly) against a failure that recurs (or is broader), so the mechanism's own success does not entail the property holding
- **first used on**: R2-3

The three fields above are the seat's own words, refused at the write if any were missing
(`record.validateClassNew`), so this proposal is well-formed by construction. What a
reviewer decides is whether it DISCRIMINATES — whether the distinguisher actually separates
it from its neighbour on a case where both are arguable.
