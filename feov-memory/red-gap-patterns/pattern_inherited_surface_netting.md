---
name: pattern-inherited-surface-netting
description: "We don't add risk, it's inherited from native/baseline" netting arguments must verify the baseline wasn't since remediated — bespoke may re-open what upstream closed
metadata:
  classes: [risk-coverage-omission, claim-contradicts-own-record]
  type: feedback
---

When blue defends a build decision with "most of the attack surface is *inherited* from the native
baseline, so adopting native instead buys no safety on this dimension," verify the baseline is still
as vulnerable as claimed.

**The trap:** the netting treats the native baseline as static-vulnerable. But if upstream *patched*
the surface (e.g. the CVE-2026-21852 fix reportedly removed user memories from the system prompt —
de-authorized the native memory surface), then post-fix native is *less* poisonable, and the bespoke
layer's high-authority projection (`.claude/rules/` loads at CLAUDE.md priority) *re-authorizes*
injection — bespoke re-widens what native narrowed. "Shared/inherited" is then false; the bespoke
layer creates net-new high-authority surface.

**Double-bind to exploit:** if the remediation detail is too low-confidence to rely on (blue itself
tagged "removed from system prompt" medium-confidence, R1-29), blue cannot also lean on it to claim
native == bespoke poisoning surface. Either direction breaks the "Shared" cell.

**Why:** the netted build-vs-adopt table is the keystone of the go/no-go; a wrong "Shared"
classification flips the conclusion.

**How to apply:** for each "inherited/shared" cell in a net-new-surface table, ask (1) was the
baseline remediated since the cited incident? (2) does the bespoke design restore authority/exposure
the remediation removed? If yes to both, reclassify as net-new.
