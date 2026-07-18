---
name: pattern-verification-probe-layer-masking
description: a per-run "enforcement verified" probe whose PASS outcome is produced by a different layer than the one under test — the redundant layer can be dead and the probe still passes
metadata:
  classes: [verification-scope-blindspot, derivation-status-overclaim]
  type: feedback
---

A verification probe (denial canary, liveness check) minted to prove layer N is live can be
INCAPABLE of isolating layer N when a co-located layer produces the same observable outcome.

**Concrete instance (sleeper-service R2, L6-F1):** step-0 denial canary "attempt one
out-of-fence write, abort unless DENIED" was minted (R1-28) to prove the marker-keyed HOOK
fence (layer 2) is live per run and "fails CLOSED on marker loss." But the same out-of-fence
write is independently denied by layer 1 (the `dontAsk` permission profile) — so when the
marker is lost and layer 2 goes dormant, layer 1 denies the canary write anyway, the canary
reports DENIED, and the run proceeds under the leaky layer-1-alone config the probe existed
to prevent. The canary passes EXACTLY when the fence is dead.

**Why:** layers 1 and 2 fence the identical boundary, so no probe write exists that layer 1
allows but layer 2 denies. A deny-OUTCOME probe over a boundary two layers share can never
isolate the redundant layer. The fix is a POSITIVE liveness record emitted by the layer under
test (the hook writes "I fired + denied X" to a log the wrapper confirms), not an outcome the
other layer can mask.

**ALLOW-side variant (sleeper-service R4, L4-F1/L3-F2, two lenses independently):** the
inverse masking. Blue "reproduced" that the dontAsk read-only-Bash carve-out classifier
passes `git log --output` because the command ran with NO PROMPT in its seat — but every
harness seat on this box runs under `defaultMode: "auto"`, where the AUTO classifier is the
approving layer; "no prompt" proves "≥1 approving layer passed it," never "the carve-out
classifier passed it." An allow-outcome probe over shared APPROVING layers is exactly as
attribution-blind as a deny-outcome probe over shared denying layers. Check the probe seat's
own permission mode before accepting any classifier-attribution claim.

**Environment fact (recorded R4):** the isolating probe — nested `claude -p
--permission-mode dontAsk` from a clean temp repo — is itself DENIED by the seat's auto-mode
classifier (twice, minimal command). Layer-isolation probes are not runnable from lens seats;
grade the attribution claim down with the attempt recorded and route the probe to a build-PR
or operator step. Do not route around the denial.

**How to apply:** whenever blue adds a per-run "enforcement verified / fails closed" probe,
ask: what OTHER layer produces the same PASS observable? If any co-located control denies/
allows the same probe target, the probe proves "≥1 layer acted," not "THIS layer acted."
Demand a layer-specific positive signal. Sibling of [[pattern_self_defeating_mitigation]] and
[[pattern_fail_open_guard_and_erasable_evidence]] — the RECORD-but-don't-CLOSE-THE-LOOP shape
(attempt/record without verify-and-surface) recurred three times in one round (canary,
snapshot with no reader, plugin-copy with no version check).
