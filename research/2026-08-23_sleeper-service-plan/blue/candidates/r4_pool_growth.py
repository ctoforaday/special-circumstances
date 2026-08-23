#!/usr/bin/env python3
"""
R4-4: bound the second, cadence-independent gate-degradation path -- rising
per-promotion SEARCH COST over the ungated research/+ideas/ pool. Model (same as
lane3_accumulation.py, restated deterministically without the live filesystem read):
per-run yield p=1 idea stub (Phase-4 verify names an 'idea stub', singular); the
pool is arrivals minus promotions; nothing archives/decays ideas/ (checked: plan
3c/Phase-4/Resolved-6 specify no ideas/ decay). unpromoted_pool_after_1yr =
(runs_per_wk - mu) * 52 for a single-maintainer mu promotions/wk.

The point R4-4 makes and this check quantifies: under DAILY automated cadence
(runs=7/wk) the pool diverges regardless of rho at the promotion gate. Under MANUAL
cadence the human sets runs_per_wk; if they run at their own promotion rate
(runs=mu) the pool is bounded BY CONSTRUCTION -- but nothing STRUCTURALLY caps it,
so any manual over-triggering, or simply not promoting every stub, still accretes.
"""
def pool_after_1yr(runs_per_wk, mu):
    return (runs_per_wk - mu) * 52

print("=== Daily automated cadence (runs=7/wk), 1yr, no decay ===")
for mu in (7, 3, 1):
    print(f"  mu={mu}/wk: unpromoted pool after 1yr = {pool_after_1yr(7, mu):4d} stubs")

print()
print("=== Manual cadence: pool bounded only by the human's trigger rate ===")
for runs in (3, 4, 5, 7):
    mu = 3
    p = pool_after_1yr(runs, mu)
    note = "bounded by construction (runs=mu)" if runs == mu else "accretes: runs>mu"
    print(f"  runs={runs}/wk, mu=3/wk: pool/yr = {p:4d}   {note}")

print()
print("VERDICT R4-4:")
print("  At mu=3/wk, daily cadence accretes ~208 unpromoted stubs/yr (312 at mu=1).")
print("  Manual cadence at runs=mu holds it at 0 by construction, but that is human")
print("  discipline, not a structural cap: no archival/decay exists for ideas/, so the")
print("  rising-search-cost path to reviewer degradation is real yet cadence-throttleable")
print("  -- a distinct disposition from the rho-row (which the manual default fully closes).")
assert pool_after_1yr(7, 3) == 208
assert pool_after_1yr(7, 1) == 312
assert pool_after_1yr(3, 3) == 0
print("  checks pass: 208 @ (daily,mu=3); 312 @ (daily,mu=1); 0 @ (manual runs=mu=3).")
