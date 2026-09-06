# proposed gap class — `unrecorded-report-edit` [PROPOSED — not in the registry until adopted]

Coined by `red-merge-r2` during `2026-09-02_quadratic-formula`. It is NOT staged into any later run:
an unreviewed class validating a future `--class` is the registry losing the only thing it
means. Adopting it means adding the slug to `feov-memory/class-registry.json` by hand.

- **definition**: The living report's text at a site differs from what the record says it is, because a change to it has no corresponding edit event on the change stack.
- **neighbour**: `claim-contradicts-own-record`
- **distinguisher**: Is the contradiction between two statements INSIDE the document (claim-contradicts-own-record), or between the document's current text and the RECORD of how that text changed (this class)?
- **first used on**: R2-8

The three fields above are the seat's own words, refused at the write if any were missing
(`record.validateClassNew`), so this proposal is well-formed by construction. What a
reviewer decides is whether it DISCRIMINATES — whether the distinguisher actually separates
it from its neighbour on a case where both are arguable.
