---
name: research-protocol
description: Use when performing or auditing deep research — the protocol (frontier hypotheses, saturation, disconfirming budget, semantic footnotes), the run-directory layout, and the debate envelopes.
---

# research-protocol

Research that survives an adversary.

## Protocol

- BEFORE searching, YOU MUST formulate 3–5 frontier hypotheses — what would be true if each candidate answer were right — and record them; searches then test hypotheses instead of wandering.
- During research, YOU MUST search to **saturation**: stop only when new searches return already-seen sources (typically 20–30 searches for a deep topic).
- During research, YOU MUST spend at least one search in five hunting **disconfirming** evidence against your current position. This is a drafting floor, not the verification: it keeps confirmation bias out of the draft; systematic disconfirmation is red's entire job.
- During writing, YOU MUST cite with semantic word-based footnotes — `[^WordLabel]` carrying title, source, and access date. Numbered footnotes are deprecated: labels stay meaningful while drafts move.
- AFTER drafting, every claim MUST trace to a source a skeptic can follow; unverifiable claims are labeled as such, not laundered into fact.
- For PDF-only sources, YOU MUST try the document-extraction MCP tools before grading down on a lossy fetch: `arxiv-latex` (exact LaTeX for arXiv figures/tables) and `pdf-reader` (page/table extraction with provenance) — discoverable via ToolSearch when the project's `.mcp.json` servers are approved. Two runs of friction ranked lossy PDF fetches the #1 capability gap; a claim capped at "unable to corroborate" without trying these is an incomplete audit.

## The run directory (the blackboard)

```
research/<date>_<slug>/
├── report.md          # final deliverable — assembled LAST, by union
├── inputs/PINNED.md   # the evidence base, pinned: repo HEAD at launch + cited corpora's commit/round
├── blue/
│   ├── frontier.md    # the hypotheses
│   ├── report.md      # blue's LIVING report — grows every round, never summarized away
│   ├── CHANGELOG.md   # what blue changed each round (keeps debate.md argument-focused)
│   └── candidates/    # best-of-N method-lens lane drafts, preserved
├── red/
│   ├── findings.md    # red's LIVING audit — cumulative verdict + graded gaps + lineage (supersedes)
│   ├── citation-ledger.md  # verified citations don't un-verify: claim | reference | confidence | round | access-date
│   └── candidates/    # per-lens audit passes (lens-scoped labels L1-F1...), preserved
├── debate.md          # the FULL three-party transcript — every round: ### RED / ### BLUE / ### LEAD (only red-merge writes ### RED)
├── friction.md        # capability/protocol complaints — seats append DURING the run (survives aborts)
├── cost.md            # measured tokens + dollars per seat-round (scripts/cost-audit.mjs)
└── trajectories/      # journal.jsonl (tracked) + agent-transcripts.tar.gz (gitignored)
```

All artifacts are git-tracked; nothing is summarized away. The payload is the file; the envelope is the handle — no large content travels through agent return values.

## Recall (qmd) — three access modes, never confused

When `qmd` is installed (optional tier; `/prosthetic-conscience:doctor` reports it), the corpus
is searchable instead of only re-readable. The discipline:

1. **Retrieval for evidence and context** — finding what the corpus says about a question.
   Swarm seats use the CLI: `qmd search "<terms>" -c <collection>` (BM25 — sub-second, zero
   models, line-anchored `qmd://` URIs citable as locations). YOU MUST NOT invoke bare-CLI
   `qmd query`/`vsearch` from a seat: each invocation reloads ~1–2GB of models (36s measured);
   model-backed search belongs to the resident MCP server (`qmd` in `.mcp.json`, via ToolSearch).
   Over MCP, YOU author the `searches` array (lex/vec/hyde sub-queries) yourself — you know the
   domain vocabulary; there is no auto-expander to mangle it.
2. **Full read for the document under audit** — red reads blue's living report whole, in
   context, every round. Retrieval NEVER substitutes: a decontextualized snippet is how audits
   go blind. This clause outranks any token saving.
3. **Leaf-node fetch for verification** — a citation is checked against its source (WebFetch,
   PDF MCPs, `qmd get <path>` for corpus-internal references), never against a search snippet.

Freshness is deterministic, not remembered: the `sc-recall-index` hook runs a fast FTS update
on every markdown write (measured ~0.7s), so BM25 is never stale. Semantic embeddings refresh
at phase tops (`qmd embed`, incremental) — at worst one phase stale; grade accordingly or
re-run `search` (lexical, always current) to confirm.

## Report structure

The final `report.md` (see `references/report_template.md`): verdict stamp (VERIFIED/UNVERIFIED + rounds) → **the Catechism** (`references/catechism_template.md` — the worth-our-time decision, adapted from Heilmeier) → analytical core (foundations / analysis / risk matrix graded likelihood × impact × complexity, including risk-accepted items with rationale) → **blue's report in full** → **red's findings in full** → debate record → **open questions carried past this run** (blue's final envelope, verbatim) → footnotes (with access dates; volatility noted for living sources).

## Friction (the complaint channel)

A subagent's only voice is its return value — so capability complaints travel in the envelope.

- AFTER any task where a missing tool, denied permission, or capability gap impeded you, YOU MUST report it in the envelope's `friction` field: name the capability and what you would have done with it.
- AFTER any task where the material did not fit the shape you were given — a template section that made no sense for the topic, a protocol step that fought the work, an envelope field you had nothing honest to put in, content with no home — YOU MUST report the misfit as friction: name the template/step/field and what shape the work actually wanted.
- YOU MUST NOT silently work around a capability gap — the workaround destroys the signal that would get you retooled.
- The lead aggregates friction into the run record (`friction.md`); the self-improvement loop consumes it. Complaints are how the system learns what its agents actually need.
