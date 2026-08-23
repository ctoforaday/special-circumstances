#!/usr/bin/env python3
"""
Q3 — the sleeper-service promotion gate as an M/M/1 queue.

Arrivals  lambda : promotion-CANDIDATE-bearing autonomous runs per week.
                   Daily default cadence => 7 runs/week; yield p in [0,1] of runs
                   that actually produce a promotable candidate => lambda = 7*p.
Service   mu     : promotions a single maintainer can review-and-clear per week.
                   Each promotion = reading a full multi-round FEOV report + judging
                   a rule/skill change (Semantic Consent gate). Bounded by human hours.

Model note: the research/ and ideas/ writes are NOT gated (cheap, unbounded, fine).
The gated resource is promotion INTO rules/skills. That is the single-server queue.
rho = lambda/mu. rho >= 1 => backlog grows without bound => the gate becomes nominal
(the human stops reading). Little's law L = lambda*W ties backlog length to wait.
"""

def rho(lam, mu):
    return lam / mu

def mm1_metrics(lam, mu):
    # standard M/M/1 steady-state (valid only for rho<1)
    r = lam / mu
    if r >= 1:
        return r, float('inf'), float('inf')
    L = r / (1 - r)          # mean number in system (Little: L = lambda*W)
    W = 1 / (mu - lam)       # mean time in system (weeks)
    return r, L, W

print("=== rho = lambda/mu grid: daily cadence (7 runs/wk), varying yield p and human throughput mu ===")
print(f"{'yield p':>8} {'lambda(=7p)':>11} | " + " ".join(f"mu={m:<4}" for m in (1,2,3,5,7)))
for p in (1.0, 0.75, 0.5, 0.25, 0.1):
    lam = 7 * p
    row = []
    for mu in (1,2,3,5,7):
        r = rho(lam, mu)
        flag = "UNSTABLE" if r >= 1 else f"{r:.2f}"
        row.append(f"{flag:>7}")
    print(f"{p:>8} {lam:>11.2f} | " + " ".join(row))

print()
print("=== Stability frontier: max yield p s.t. rho<1, per human throughput mu (daily cadence) ===")
for mu in (1,2,3,5,7):
    # 7p < mu  => p < mu/7
    p_star = mu / 7.0
    print(f"  mu={mu}/wk  -> stable only if candidate-yield p < {min(p_star,1.0):.3f}"
          + ("  (always stable at daily cadence)" if p_star>=1 else ""))

print()
print("=== Backlog blow-up near the frontier (M/M/1, Little's law L=lambda*W), mu=3/wk ===")
mu = 3.0
for p in (0.1, 0.25, 0.40, 0.42, 0.428):
    lam = 7*p
    r, L, W = mm1_metrics(lam, mu)
    Ls = "inf" if L==float('inf') else f"{L:6.1f}"
    Ws = "inf" if W==float('inf') else f"{W:5.2f}"
    print(f"  p={p:<5} lambda={lam:4.2f}/wk rho={r:4.2f}  mean backlog L={Ls} reports  mean wait W={Ws} wk")

print()
print("=== Sanity: pure daily cadence with EVERY run promotable (p=1, lambda=7/wk) ===")
for mu in (1,2,3,5,7):
    r,_,_ = mm1_metrics(7.0, mu)
    print(f"  mu={mu}/wk -> rho={7.0/mu:.2f} " + ("(UNSTABLE, backlog->inf)" if r>=1 else "(stable)"))
