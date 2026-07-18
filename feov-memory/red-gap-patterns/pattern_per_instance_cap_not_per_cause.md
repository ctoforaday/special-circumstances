---
name: pattern-per-instance-cap-not-per-cause
description: A retry/waste cap bounds the INSTANCE (run-dir, item, attempt) while the generating CAUSE freely mints new instances — "cannot recur forever" claims that hold per-instance only
metadata:
  classes: [recurrence-detector-keying, false-universal]
  type: feedback
---

When a design closes a livelock/waste gap with a bounded retry cap (k attempts per X),
check what X is keyed on. If the cap is per-INSTANCE (per run-dir, per item, per file)
and the failure's root CAUSE survives instance replacement (deterministic abort, corrupt
shared input, persistent dependency outage), the mechanism re-opens the bounded waste on
every fresh instance: cap trips → fresh instance minted → same cause re-wedges it →
repeat until an outer budget trips, then the outer budget RESETS (monthly) and the burn
resumes. The sentence often self-incriminates with an escape clause ("...until the
monthly cap trips") while headlining "cannot happen forever."

**Why:** sleeper-service round 2 (R2-10, superseding R1-29): resume cap k=3 per run-dir
+ "next fire mints a fresh dir" = ~k×$5/night with zero output for any deterministic
root cause, ~3 nights to the $50 monthly cap, resetting every month — the livelock the
fix claimed to close, relocated one level up.

**How to apply:** for every bounded-retry fix, ask "what happens on the N+1th INSTANCE
of the same cause?" Demand a per-cause signature (hash of abort reason/failure class)
with a halt-and-surface threshold, or an honest re-scope of the claim to its per-instance
truth. Sibling of [[pattern-self-defeating-mitigation]] (control's own failure mode) and
the controller-lookahead/stale-baseline pricing family (mechanism priced on the state it
cannot see).
