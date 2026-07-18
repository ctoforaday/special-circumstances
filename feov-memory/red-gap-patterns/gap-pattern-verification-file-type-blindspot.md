---
name: gap-pattern-verification-file-type-blindspot
description: Gap pattern — "no X exists" claims backed by a grep scoped to one file type miss X implemented in another layer (e.g. compiled tools)
metadata:
  classes: [verification-scope-blindspot, false-universal]
  type: feedback
---

Gap pattern to hunt: a blue "no such thing exists / must be built from scratch" claim whose
supporting local verification is a grep scoped to a **single file type**.

**Why:** In the 2026-07-12 memory-architecture audit, blue asserted "no secret-scrub gate exists,
build it" (§6.3, §8 item 3) backed by footnote `[^LocalRepoScrub]`: `grep -i secret|scrub|denylist
across *.md`. That scope was blind to the Go tool layer. A shipping PreToolUse hook
(`sc-secrets-gate`), a reusable pattern package (`internal/secrets`), tests, and `hooks.json`
wiring already existed. The claim miscast a blocking item's effort ("build from zero" vs "wire a
new consumer onto an existing package").

**How to apply:** When a report says "X does not exist" and cites a local grep, check the grep's
`--include`/glob scope. Non-existence claims are only as strong as the search surface. Re-run
across code (`*.go`, `*.ts`, `*.py`), config (`*.json`, `*.toml`), and wiring/manifest files
before accepting. A capability can be present in a compiled binary or hook manifest while absent
from the prose/skill markdown that describes it (which may lag in future tense). Also check the
narrower truth: the existing capability may cover a *different surface* than the design needs
(here: outbound-tool-input scan, not commit/push-time scan) — that nuance is the real gap.

**Compounding form (Round 1 merge, same audit):** a *second* independent verifier (a different red
lens) repeated the identical `*.md`-scoped grep and **corroborated the false "no gate exists" claim
HIGH**. Two verifiers agreeing does NOT raise confidence when they share the same flawed method
scope — the agreement is an artifact of the shared blindspot, not independent confirmation. When
consolidating multiple lens/verifier passes, do not treat concurrence as corroboration until you
confirm the verifiers used *different* search surfaces. Re-verify shared-method agreements at the
leaf node yourself (I did; found `sc-secrets-gate` + `internal/secrets` + `hooks.json` live).
