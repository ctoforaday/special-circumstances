# proposed gap class — `guardrail-scope-overclaim` [PROPOSED — not in the registry until adopted]

Coined by `red-merge-r1` during `2026-08-23_sleeper-service-plan`. It is NOT staged into any later run:
an unreviewed class validating a future `--class` is the registry losing the only thing it
means. Adopting it means adding the slug to `feov-memory/class-registry.json` by hand.

- **definition**: a safety/tolerability argument states a control (e.g. write-confinement) makes a set of otherwise-open risks safe to defer, without stating the actual scope of what the control covers, so a reader accepts blanket coverage the control does not have
- **neighbour**: `policy-without-mechanism`
- **distinguisher**: policy-without-mechanism is a control that exists only as prose with no enforcing binary at all; guardrail-scope-overclaim is a control that DOES have (or will have) a mechanism, but the report claims that mechanism covers more risk surface than it actually does
- **first used on**: R1-1

The three fields above are the seat's own words, refused at the write if any were missing
(`record.validateClassNew`), so this proposal is well-formed by construction. What a
reviewer decides is whether it DISCRIMINATES — whether the distinguisher actually separates
it from its neighbour on a case where both are arguable.
