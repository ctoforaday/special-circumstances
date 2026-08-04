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
- During writing, YOU MUST add every citation with the TOOL — `blue cite --location "<the exact sentence>" --url <u> --title <t>` — never by hand. The tool fetches the source once into the run cache, then splices an INVISIBLE, IMMORTAL `<!--cite:c-…-->` anchor at that sentence; assembly weaves the anchors into the visible `[^N]` footnotes and composes the `## Bibliography`. A hand-typed `[^label]` is not a citation: nothing backs it, the claim counter does not see it, and the unbacked-citations detector flags it. An unreachable source is unusable — the cite is rejected and logged as friction.
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
│   ├── ledger.md      # SINGLE SOURCE OF TRUTH for status: open gaps (full grading) + compact
│   │                  #   closure index (id | class | one-line summary | supersedes) — red-merge-born
│   │                  #   round 1 (write-guard-verified names), NOT skeleton-born
│   ├── archive.md     # immutable closed prose, append-only; read on demand (near-match, chain
│   │                  #   rulings, spot-checks) — never resident in the default merge/judge read
│   └── citation-ledger.md  # verified citations don't un-verify: claim | reference | confidence | round | access-date
│                      #   (lens findings are RECORD EVENTS now — read via `show --view findings`, no candidate files)
├── debate.md          # the FULL three-party transcript — every round: ### RED / ### RED CLOSING /
│                      #   ### BLUE / ### BLUE CLOSING / ### LEAD (adjudication sits LAST: the judge
│                      #   rules on the closings, the transcript, and the final artifact state only)
│                      #   friction lives on the RECORD (the friction verb; read via `show --view friction`),
│                      #   not a file — the pre-tool friction.md was retired 2026-07-19
├── cost.md            # measured tokens + dollars per seat-round (feov-record cost)
└── trajectories/      # journal.jsonl (tracked) + board-telemetry.jsonl (one JSON line per round:
                       #   board profile, mass under the pinned mapping, accepted-dispute deltas —
                       #   the SIGNAL the stopping judgment reads; convenience copy, never the
                       #   evidence of record) + agent-transcripts.tar.gz (gitignored)
```

**Termination is judged, and the standing practice is stop-and-resume**: `maxRounds` is a cost
ceiling, never the terminator of record. The operator reads the board telemetry (open count,
max severity, new-mint profile, mass trend), stops a run past its value, and resumes with a
reduced `maxRounds` for the honest UNVERIFIED assembly — cache replay makes the stop ~$0
(measured). Automatic severity-floor termination was evaluated and REJECTED (run-4 report §1):
it automates the one call that belongs to judgment. NEVER change models on the resume.

All artifacts are git-tracked; nothing is summarized away. The payload is the file; the envelope is the handle — no large content travels through agent return values.

## Recall (qmd) — three access modes, never confused

When qmd is installed (`/prosthetic-conscience:doctor` installs the pinned version on consent)
and the project's `.mcp.json` declares the `qmd` server, the corpus is searchable instead of
only re-readable. Exactly ONE qmd exists — the installed binary; the MCP server and the
`sc-recall-index` hook both run it, so read and write paths can never skew on index schema.
ALL seat access goes through the MCP server (tools `query`/`get`/`multi_get`/`status`,
discoverable via ToolSearch) — the resident process loads models once; the bare CLI reloads
~1–2GB per invocation (36s measured) and MUST NOT be used for search from any seat. The
discipline:

1. **Retrieval for evidence and context** — finding what the corpus says about a question.
   Use the MCP `query` tool and author the `searches` array yourself — you know the domain
   vocabulary; there is no auto-expander to mangle it. Cheap and lexical: a single
   `{type: "lex", query: "<terms>"}` sub-query (BM25, line-anchored `qmd://` URIs citable as
   locations). Semantic, when phrasing won't match wording: add `vec` (rephrasings) and `hyde`
   (a hypothetical passage that would answer you) sub-queries.
2. **Full read for the document under audit** — red reads blue's living report whole, in
   context, every round. Retrieval NEVER substitutes: a decontextualized snippet is how audits
   go blind. This clause outranks any token saving.
3. **Leaf-node fetch for verification** — a citation is checked against its source, never against a
   search snippet. For a source BLUE CITED, read the exact bytes blue read from the run cache
   (`fetch --url <the cited url>` — a cache hit, so you audit the same artifact, not a page that may
   have drifted since). For a source you discover yourself, pull it verbatim (Bash `curl`, PDF MCPs,
   MCP `get` for corpus-internal references). WebFetch is not used: it returns a summary, not the source.

Freshness is deterministic, not remembered: the `sc-recall-index` hook runs a fast FTS update
on every markdown write (measured ~0.7s), so lexical results are never stale. Semantic
embeddings refresh at phase tops (incremental) — at worst one phase stale; grade accordingly
or confirm with a lex-only query (always current). Index maintenance (collections, update,
embed) is lead/hook mechanics, never seat work.

## Harness contract (one referenceable paragraph — three seats re-derived this at token cost)

The Workflow script's `log()` is operator-console-EPHEMERAL: it persists nowhere. The
transcript directory's `journal.jsonl` is the HARNESS's lifecycle record — `started`/`result`
events only, never script logs. Per-agent API transcripts are `agent-*.jsonl` (the cost
audit's input). Durable in-run state lives ONLY on the blackboard (git-tracked run files) or
in envelopes; anything else evaporates with the session. Tool footguns with live recurrences:
Grep's count mode counts LINES, not occurrences (anchor patterns when counting); quoted
heredocs can eat backslashes (prefer the Write tool for scripts); the Read tool caps ~25k
tokens — a full-document read over that cap is consecutive whole windows, which satisfies the
full-re-read MUST without a confidence discount.

## Report structure

The final `report.md` (see `references/report_template.md`): verdict stamp (VERIFIED/UNVERIFIED + rounds) → **the Catechism** (`references/catechism_template.md` — the worth-our-time decision, adapted from Heilmeier) → analytical core (foundations / analysis / risk matrix graded likelihood × impact × complexity, including risk-accepted items with rationale) → **blue's report in full** → **red's findings in full** → debate record → **open questions carried past this run** (blue's final envelope, verbatim) → footnotes (with access dates; volatility noted for living sources).

## Friction (the complaint channel)

A subagent's only voice is its return value — so capability complaints travel in the envelope.

- AFTER any task where a missing tool, denied permission, or capability gap impeded you, YOU MUST report it in the envelope's `friction` field: name the capability and what you would have done with it.
- AFTER any task where the material did not fit the shape you were given — a template section that made no sense for the topic, a protocol step that fought the work, an envelope field you had nothing honest to put in, content with no home — YOU MUST report the misfit as friction: name the template/step/field and what shape the work actually wanted.
- YOU MUST NOT silently work around a capability gap — the workaround destroys the signal that would get you retooled.
- Friction is recorded on the RECORD via the friction verb (read via `show --view friction`); capture reconciles every envelope's `friction` field against it, and the self-improvement loop consumes it. Complaints are how the system learns what its agents actually need.
