---
name: pattern-anchor-not-in-sources
description: Before flagging a report cite anchor as unbacked, check the evidence projection's `independent` list, not just `sources`
metadata:
  type: feedback
---

When diffing report `cite:` anchors against the evidence projection to hunt unbacked citations, DO NOT conclude "anchor in report but not in ledger => unbacked" from the `sources` array alone.

**Why:** The evidence projection has multiple keys — `sources`, `independent`, `proofs`, `reopened`, `unanswered_contradictions`, `counts`. A red CORROBORATION (red went and found a source for a claim blue made, then blue added a `cite` label) lands in `independent` carrying a `label: c-xxxx`, NOT as a top-level `sources` object. A naive `set(report_anchors) - set(s.anchor for s in sources)` produces FALSE unbacked-citation positives for every such anchor. Confirmed in the 2026-09-02 quadratic run: 6 report anchors were "missing" from `sources` but 5 were labeled entries in `independent` (Crossref/ERIC/IA corroborations) and 1 was a documented stranded title anchor.

**How to apply:** When an anchor appears in the report but not in `sources`, walk the whole evidence structure for that anchor string (check `independent[].label` and `sources[].verified[].text`) before minting a `finding`. Cross-check against `counts.sources_unverified` — if it is 0, every source object is verified and any true gap is a stranded/orphaned anchor, which is usually already docketed (check the closed board for `misplaced-citation-anchor`).
