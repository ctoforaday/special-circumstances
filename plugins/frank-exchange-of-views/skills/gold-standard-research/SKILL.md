---
name: gold-standard-research
description: Use when performing or auditing deep research — the protocol (frontier hypotheses, saturation, disconfirming budget, semantic footnotes), the run-directory layout, and the debate envelopes.
---

# gold-standard-research

Research that survives an adversary.

## Protocol

- BEFORE searching, YOU MUST formulate 3–5 frontier hypotheses — what would be true if each candidate answer were right — and record them; searches then test hypotheses instead of wandering.
- During research, YOU MUST search to **saturation**: stop only when new searches return already-seen sources (typically 20–30 searches for a deep topic).
- During research, YOU MUST spend at least one search in five hunting **disconfirming** evidence against your current position.
- During writing, YOU MUST cite with semantic word-based footnotes — `[^WordLabel]` carrying title, source, and access date. Numbered footnotes are deprecated: labels stay meaningful while drafts move.
- AFTER drafting, every claim MUST trace to a source a skeptic can follow; unverifiable claims are labeled as such, not laundered into fact.

## The run directory (the blackboard)

```
research/<date>_<slug>/
├── report.md          # final deliverable — assembled LAST, by union
├── blue/
│   ├── frontier.md    # the hypotheses
│   ├── report.md      # blue's LIVING report — grows every round, never summarized away
│   ├── CHANGELOG.md   # what blue changed each round (keeps debate.md argument-focused)
│   └── candidates/    # best-of-N lane drafts, preserved
├── red/
│   ├── findings.md    # red's LIVING audit — cumulative verdict + graded gaps
│   └── candidates/    # per-lens audit passes, preserved
└── debate.md          # the FULL three-party transcript — every round: ### RED / ### BLUE / ### LEAD
```

All artifacts are git-tracked; nothing is summarized away. The payload is the file; the envelope is the handle — no large content travels through agent return values.

## Report structure

The final `report.md` (see `references/report_template.md`): verdict stamp (VERIFIED/UNVERIFIED + rounds) → **Heilmeier Catechism** (`references/heilmeier_template.md`) → analytical core (foundations / analysis / risk matrix graded likelihood × impact × complexity, including risk-accepted items with rationale) → **blue's report in full** → **red's findings in full** → debate record → footnotes.
