---
name: pattern-invariant-soundness-by-enumeration
description: A keystone invariant sold as "sound/mechanical" whose soundness actually rests on an under-inclusive channel denylist — provably incomplete when the system's own symmetric defense already treats an omitted channel as I/O
metadata:
  type: feedback
---

When blue closes a cluster of gate-by-gate patches by adopting a single **invariant** (the
anti-complexity move red often recommends), the invariant is frequently claimed **"sound"** or
**"mechanical"** — but its soundness rests on an **enumerated denylist** of the channels it must
cover, and the enumeration has holes.

**Why:** an invariant like "external-touched ⇒ tainted, transitively" is only sound if "external"
is the *complete* set of channels through which attacker-authored bytes enter. Blue tends to fix a
short denylist (e.g. `WebFetch`/`WebSearch`/`file:`/`ingest`) and omit routine channels: `Bash`-fetched
bytes (`curl`/`gh`/`git log`), MCP tool results, sub-agent **sidechain** reads that launder into the
parent, and in-repo files authored by untrusted commits read via `Read` (not classed "external").
A denylist is the wrong structure for a taint/trust boundary — a newly-added tool defaults to
*trusted*.

**The killer tell:** the system's *own symmetric defense* already treats an omitted channel as I/O.
In the memory-architecture debate (R4-1), blue's outbound secret-gate wired on `WebFetch|WebSearch|Bash`
— proving `Bash` is a first-class channel — while the inbound taint invariant omitted `Bash`. The
exfil pipe and the injection pipe are the same pipe; defending one end and not the other is the gap.
Look for this asymmetry to prove incompleteness at the leaf node rather than merely asserting it.

**How to apply:**
- When an invariant is claimed "sound"/"mechanical"/"by construction," find its channel/field
  enumeration and test each *routine* channel the enumeration omits.
- Recommend **inverting to an allowlist** (enumerate what is *trusted*; everything else taints) so
  new channels default safe — this is usually a parser change, not research, so grade it a hardening
  gap, not a redesign.
- Related: an invariant can also **name** a field in its statement but **omit** it from the
  corollary's reset/enforcement list (R4-4: `last_seen` named non-inheritable but not reset on
  import) — verify the mechanism executes every leg the invariant claims, same discipline as
  [[pattern_incomplete_repair_footnote_lag]] (retract-by-annotation vs actual edit).
- Every downstream closure that leans on the invariant ("closed by construction under §X") is
  **contingent** on the enumeration being complete — surface the contingency, don't accept the
  cascade of closures on an unproven root. Relates to [[pattern_missing_root_invariant]] (the prior
  round's recommendation) and [[pattern_risk_grading_conflations]] (verdict/docket keystone leaning
  on an unverified mechanism).
- **Two enumeration-completeness flavors caught in the sleeper-service run (R4-2/R4-3):**
  (1) **Source-hedged list.** "Deny-enumerated per command" was built from a doc that names the
  read-only set with "These **include** …" and never claims completeness — the same page names
  `sort`/`sed` as classifier-reasoned commands NOT in the enumerated list. When a closure's
  enumeration is copied from a source that itself hedges ("include", "such as", "e.g."), the
  closure inherits the hedge; grep the source sentence for a totality word before trusting a
  "fully enumerated" claim.
  (2) **Functional-exempt member.** A "deny-enumerated per command" carve-out RETAINED one member
  un-enumerated because the design NEEDS it (read-only git) — and that exact member was the one
  with a leaf-confirmed write gadget (`git format-patch -o` → out-of-repo file). An enumeration
  that intentionally exempts a member for functional need is not "fully enumerated"; the exempt
  member is where the next gadget lives (sibling-halo — see [[pattern_audited_artifact_sibling_halo]]
  and [[pattern_wildcard_allow_write_gadget]]). The allowlist-inversion recommendation applies to
  the exempt member specifically: allowlist the exact read forms, deny the rest.
  (3) **Lexical-form escape recurring INSIDE the repair (R5-5).** A belt extension minted to close
  a sibling escape (`Bash(* -o *)`, added to catch the short-form output flag the `--output` denies
  missed) is itself escaped by the NEXT lexical form: `Bash(* -o *)` requires a space, but getopt
  short options accept the ATTACHED form `-o<value>` (`git format-patch -1 -o/tmp/x` → exit 0,
  out-of-repo — leaf-reproduced at lens AND merge seats). The enumerate-and-extend approach the
  belt keeps doubling down on cannot close the class — each new pattern is escaped one lexical form
  deeper. The fix is NOT another belt pattern; it is to confirm the allowlist (hook parses `git`
  argv, rejects `format-patch` entirely) handles attached-form flags, and add the attached form to
  the probe matrix so it is leaf-tested, not assumed. When you catch this, re-run the gadget in the
  ATTACHED form specifically — a spaced-form-only reproduction understates the escape surface.
