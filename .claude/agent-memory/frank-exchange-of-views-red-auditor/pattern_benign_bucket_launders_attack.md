---
name: pattern-benign-bucket-launders-attack
description: A fix that declares a denial-stream "expected/normal" keys the benign bucket on the protected TARGET path, so an attacker aiming at that target inherits the benign label — the signal is erased, not distinguished
metadata:
  type: feedback
---

When a prior round asks "are routine denials distinguishable from attack evidence in the
fired-record?", a fix that ANSWERS by bucketing all denials to that target as a NORMAL class
(`expected-memory-deny`, doctor line does not flag) does not distinguish — it ERASES the
signal. If the bucket is keyed on the TARGET PATH (`.claude/agent-memory/**`) rather than the
ORIGINATING ACTOR/CAUSE, then any actor targeting that path — including an injection trying to
poison the protected surface — is auto-classified benign. The write stays blocked (enforcement
intact) but DETECTION is gone, on precisely the surface tamper-attempts most need to be visible.

**Why:** round-5 sleeper-service R4-4. The nightly red seat's own memory write is denied by
design (correct); to keep the resulting denial-stream un-alarming, the parser buckets ALL
`.claude/agent-memory/**` denials as `expected-memory-deny`. That surface is red's gap-pattern
corpus (steers future audits) — a write attempt there is as alarming as a guardrail delta, but
it's routed to the one bucket the doctor line ignores. Mirrors the origin-tag lesson (tag by
CAUSE not name) inverted onto a denylist bucket.

**How to apply:** when a fix makes a security-relevant event "expected/normal/routine," check
what KEY the benign classification uses. Target-path-keyed benign buckets are attacker-inheritable
(the attacker chooses the target). Demand the benign label be keyed on the ORIGINATING ACTOR +
the specific expected write shape; every OTHER actor hitting that target surfaces as a distinct
persistent anomaly. "Enforcement intact but detection erased" is still a gap — grade it on lost
attack visibility, not on breach (the write is blocked).
