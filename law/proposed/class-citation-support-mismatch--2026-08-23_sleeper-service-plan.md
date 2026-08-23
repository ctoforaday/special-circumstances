# proposed gap class — `citation-support-mismatch` [PROPOSED — not in the registry until adopted]

Coined by `red-merge-r1` during `2026-08-23_sleeper-service-plan`. It is NOT staged into any later run:
an unreviewed class validating a future `--class` is the registry losing the only thing it
means. Adopting it means adding the slug to `feov-memory/class-registry.json` by hand.

- **definition**: a cited source is the correct/intended paper, but its verified content does not establish the specific theorem or figure the report attributes to it
- **neighbour**: `citation-figure-misattribution`
- **distinguisher**: citation-figure-misattribution is the WRONG source (different paper/authors/year) cited for the bytes actually fetched; citation-support-mismatch is the RIGHT source, whose content simply does not establish the specific claim attached to it
- **first used on**: R1-9

The three fields above are the seat's own words, refused at the write if any were missing
(`record.validateClassNew`), so this proposal is well-formed by construction. What a
reviewer decides is whether it DISCRIMINATES — whether the distinguisher actually separates
it from its neighbour on a case where both are arguable.
