# proposed gap class — `risk-taxonomy-conflation` [PROPOSED — not in the registry until adopted]

Coined by `red-merge-r3` during `2026-08-23_research-loop-counterparts`. It is NOT staged into any later run:
an unreviewed class validating a future `--class` is the registry losing the only thing it
means. Adopting it means adding the slug to `feov-memory/class-registry.json` by hand.

- **definition**: Two or more distinct attack/risk categories are folded under one label, producing a single no-control-covers-this conclusion that undercounts how many controls are actually missing.
- **neighbour**: `risk-coverage-omission`
- **distinguisher**: risk-coverage-omission is one named vector left uncontrolled; this class is when the omission conclusion itself rests on merging two vectors that need different controls -- ask: would splitting the label change how many missing controls the reader counts?
- **first used on**: R3-3

The three fields above are the seat's own words, refused at the write if any were missing
(`record.validateClassNew`), so this proposal is well-formed by construction. What a
reviewer decides is whether it DISCRIMINATES — whether the distinguisher actually separates
it from its neighbour on a case where both are arguable.
