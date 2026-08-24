# proposed gap class — `confinement-scope-gap` [PROPOSED — not in the registry until adopted]

Coined by `red-merge-r3` during `2026-08-23_sleeper-service-plan`. It is NOT staged into any later run:
an unreviewed class validating a future `--class` is the registry losing the only thing it
means. Adopting it means adding the slug to `feov-memory/class-registry.json` by hand.

- **definition**: A keystone safety claim is modeled over only one write channel (e.g. local filesystem paths) while the process's actual write surface spans additional, structurally different channels (network egress, external API/MCP write verbs, git remote operations) the model never examines, so the claimed guarantee is over-read relative to what the named mechanism restricts.
- **neighbour**: `mitigation-coverage-mismatch`
- **distinguisher**: mitigation-coverage-mismatch (R2-3) is a TEMPORAL gap -- a control that covers a failure at one point in time (boot) but not its recurrence (every cycle); confinement-scope-gap is a CHANNEL/SURFACE gap -- a control that covers one channel (filesystem) while a structurally different channel (network/API) is never modeled, regardless of timing.
- **first used on**: R3-4

The three fields above are the seat's own words, refused at the write if any were missing
(`record.validateClassNew`), so this proposal is well-formed by construction. What a
reviewer decides is whether it DISCRIMINATES — whether the distinguisher actually separates
it from its neighbour on a case where both are arguable.
