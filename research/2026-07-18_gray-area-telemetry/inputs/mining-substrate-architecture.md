## Architecture for the mining substrate (operator, 2026-07-18)

Two mechanisms, one available now and one the real answer.

### Now: spend an agent to hold the corpus, then query it live

A subagent reads the trajectory once into its own context and is then queried repeatedly via `SendMessage`. No build required, works today, and it keeps a 6MB corpus out of the lead's context entirely — the same context-shield pattern already used for the classification pass and the friction sweep.

### Later: an MCP that persists across operations and caches the tree in queryable form

The durable answer. Precedent already exists in-repo: **qmd is exactly this shape** — a persistent index plus an MCP server, with its lessons already paid for (per-user index location, collections config with absolute paths, staleness and update discipline, and the caching trap where a stale index answers confidently).

The natural schema is the one the probe found: events keyed by `uuid`, parented by `parentUuid`, carrying `timestamp`, `agentId`, `attributionAgent`, `effort`, and — for tool calls — name, input, and the linked `tool_result`. Queries the consumers actually want:

- every tool call by seat X touching target Y *(bench integrity: claim vs act)*
- rework — the same tool and target repeated *(seats going in circles)*
- gaps and stalls between timestamps *(wall-clock forensics, currently by hand)*
- user messages and their surrounding context *(human frustration)*
- steps between two points in the chain *(what actually happened between the claim and the closure)*

### The distinction that matters: exploration vs adjudication

These two mechanisms are **not interchangeable**, and the deception use-case is exactly where conflating them would hurt.

**An agent-as-index is a summarizer.** Its answers are non-deterministic, unreproducible, and unciteable — two identical questions can get two different answers, and neither can be checked. For *hypothesis generation* ("what patterns are in here?", "does this seat look like it went in circles?") that is fine and very cheap.

**For a FINDING it is disqualifying.** We have spent this whole cycle removing self-report from the evidence chain — the attestation audit exists precisely because "the seat says it verified" is not evidence. Replacing that with "an agent says the transcript shows" reintroduces the same defect one layer up, and it would be harder to spot because the summarizer is on our side.

So the rule for Gray Area:

> **Exploration may summarize. Adjudication must cite.**

Any query backing a bench finding must return the primary evidence — the `uuid`, the line, the tool call — so the finding cites the trajectory rather than the index's opinion of it. This is the citation-ledger discipline applied to our own tooling: verified at the leaf, or not verified.

### Consequence for the MCP design

The index becomes a trusted component, and a trusted component that can lie silently is the worst kind. Two properties follow:

- **Every answer carries its provenance** (uuid / line offset / file), so a consumer can always drop to the raw trajectory and check.
- **Staleness must fail loudly, not quietly.** qmd's cache taught this the hard way in this repo — and so did our own golden runner, which reported "recorded" while Go's test cache meant it had written nothing. An index that answers confidently from stale data is the same failure with more leverage.
