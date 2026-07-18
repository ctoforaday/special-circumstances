---
name: pattern-reflexivity-blindspot
description: report headlines a risk class for its subject matter but never applies it to its own toolchain/pipeline/drafting window — audit the report's own operations against its own findings
metadata:
  classes: [reflexivity-blindspot, risk-coverage-omission]
  type: feedback
---

A report that establishes a risk class as a headline finding about its *subject* often exempts
its *own operations* from that same class — silently, not by argued out-of-scope.

**Why:** caught 3 co-occurring instances in one round (FEOV retrospective, 2026-07-13, R1-14/
R1-15 + the R1-1 root): (a) the report treated CVE-class supply-chain poisoning as load-bearing
evidence about a sibling run, then graded adopting two third-party MCP servers into its own
citation-verification path on cost alone — no pin/review/permission-scoping line; (b) the same
poisoning class (untrusted fetched content -> agent context -> downstream action) was never
asked of FEOV's own WebFetch/WebSearch research phase, across all 18 graded risk rows; (c) the
report distrusted the corpus's self-reported status ("a backlog checkbox is not a diff") and
taught the live-source-drift lesson for external citations, but applied neither to its own
drafting window — its keystone "unmerged" claim went stale 8 minutes after verification with no
pinned-SHA/re-verify-before-acting discipline. Related: [[pattern-self-referential-repo-drift]]
(instance c is that pattern's root cause), [[pattern-doctrine-vs-implementation]].

**How to apply:** for every headline risk class or methodological rule the report asserts, ask
"does the report's own pipeline/tooling/build process sit inside this class, and does the text
either grade it or explicitly risk-accept it?" Silence is the gap. Close by addition (one graded
row or a one-line argued out-of-scope), not by blocking the underlying recommendation —
proportionality still applies.
