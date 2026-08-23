# proposed gap class — `mitigation-self-contradiction` [PROPOSED — not in the registry until adopted]

Coined by `red-merge-r4` during `2026-08-23_sleeper-service-plan`. It is NOT staged into any later run:
an unreviewed class validating a future `--class` is the registry losing the only thing it
means. Adopting it means adding the slug to `feov-memory/class-registry.json` by hand.

- **definition**: A stated mitigation is logically incompatible with another requirement or deliverable the same document states, not merely absent from an axis of risk
- **neighbour**: `mitigation-coverage-mismatch`
- **distinguisher**: coverage-mismatch is a mitigation that fails to REACH a risk axis it is supposed to cover (a gap); self-contradiction is a mitigation whose own stated scope CONFLICTS with another stated requirement (the mitigation and the requirement cannot both hold)
- **first used on**: R4-2

The three fields above are the seat's own words, refused at the write if any were missing
(`record.validateClassNew`), so this proposal is well-formed by construction. What a
reviewer decides is whether it DISCRIMINATES — whether the distinguisher actually separates
it from its neighbour on a case where both are arguable.
