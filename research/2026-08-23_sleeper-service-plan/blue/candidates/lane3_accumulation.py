#!/usr/bin/env python3
# lane-3: reframe Q3 (cadence) at the leaf. Lane 2 modelled a BLOCKED server queue
# (rho=lambda/mu) at the promotion gate. But the plan's guardrail makes research/+ideas/
# writes UNGATED (they accumulate freely) -- so daily cadence does not build an unstable
# SERVER queue; it builds a monotonically GROWING candidate POOL that the human must
# search to choose a promotion from. The cost that diverges is SELECTION over the pool,
# not wait time in a queue. This complements lane 2, it does not restate it.
#
# Grounding at the leaf:
#  - Phase-4 verify names the per-run yield: "produces a run dir + idea stub" (singular)
#    => deterministic p ~= 1 candidate/run (not probabilistic).
#  - Resolved decision 6: daily default => 7 runs/wk.
#  - No decay/archival policy for ideas/ is specified in the plan (checked: 3c, Phase 4,
#    Resolved decisions). Absent decay, the pool is strictly increasing.
import os, glob

def corpus_now(root="."):
    files, bytes_ = 0, 0
    for d in ("ideas", "research"):
        for p in glob.glob(os.path.join(root, d, "**", "*"), recursive=True):
            if os.path.isfile(p):
                files += 1; bytes_ += os.path.getsize(p)
    return files, bytes_

f, b = corpus_now(".")
print(f"current ungated corpus (ideas/+research/): {f} files, {b} bytes")
print("  -> the store already accumulates; the plan's 'empty scaffolding on publish' is stale.\n")

per_run = 1           # idea stub / run (Phase-4 verify, singular)
runs_wk = 7           # daily (Resolved decision 6)
print("candidate-pool size vs human promotion rate mu (promotions/wk), no decay, 1yr:")
arrivals_yr = per_run * runs_wk * 52
print(f"  arrivals/yr = {arrivals_yr} idea stubs")
for mu in (7, 3, 1):
    promoted = mu * 52
    residual = arrivals_yr - promoted
    print(f"  mu={mu}/wk: promoted/yr={promoted:4d}  unpromoted pool after 1yr={residual:4d}"
          f"   selection set the human searches each cycle grows to ~{residual}")
print()
print("robust claim (decay-free): pool(t) = (7 - mu_wk)*52*t grows without bound for any")
print("single-maintainer mu<7/wk; the diverging cost is SELECTION over the pool (Q2's")
print("ranking problem), monotonically harder each cycle. Manual/pull cadence sets")
print("arrivals = the rate the human actually promotes, holding the pool bounded by")
print("construction. This is the same resolution lane 2 reached (pace by the bottleneck),")
print("reached from the ungated-store side rather than the blocked-queue side.")
