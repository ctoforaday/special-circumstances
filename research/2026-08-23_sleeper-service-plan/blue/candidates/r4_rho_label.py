#!/usr/bin/env python3
"""
R4-5: which utilization rho actually backs the finite illustrative pair the report
labels 'at rho=1.00'? Method: M/M/1 steady state, L = rho/(1-rho); Little's law
L = lambda*W with lambda = rho*mu (Little's-law method cited c-c6032c0e). An M/M/1
queue at EXACTLY rho=1 is null-recurrent: L and W are infinite, so any FINITE pair
cannot sit at rho=1 -- it sits just below. Exact arithmetic via fractions.
"""
from fractions import Fraction as F

mu = F(3)  # promotions/week (documented illustrative service rate)

def rho_from_L(L):          # invert L = rho/(1-rho)  ->  rho = L/(L+1)
    return F(L) / (F(L) + 1)

def W_weeks(rho):           # time in system = 1/(mu - lambda), lambda = rho*mu
    lam = rho * mu
    return F(1) / (mu - lam)

for label, L in (("~14 items / ~5 wk  (report labels rho=0.93)", 14),
                 ("~749 items / ~250 wk (report labels rho=1.00)", 749)):
    rho = rho_from_L(L)
    W = W_weeks(rho)
    print(f"{label}")
    print(f"   exact rho = {L}/{L+1} = {float(rho):.6f}   (2dp -> {float(rho):.2f}, 3dp -> {float(rho):.3f})")
    print(f"   implied W = {float(W):.3f} wk   lambda = {float(rho*mu):.4f}/wk   mu-lambda = {float(mu-rho*mu):.4f}/wk")
    print(f"   is rho exactly 1.00? {rho == 1}")
    print()

rhoB = rho_from_L(749)
print("VERDICT R4-5:")
print(f"  The 749/250 pair sits at rho = 749/750 = {float(rhoB):.6f} (~0.999, 'just below 1'),")
print(f"  NOT rho=1.00. At rho=1 exactly, L=rho/(1-rho) diverges (infinite backlog/wait),")
print(f"  so a finite pair labelled 'rho=1.00' collides with the same paragraph's own")
print(f"  'grows without bound once rho>=1'. It only rounds to 1.00 at two decimals.")
assert rhoB != 1
assert round(float(rhoB), 2) == 1.00 and round(float(rhoB), 3) == 0.999
print("  checks pass: rounds to 1.00 at 2dp, 0.999 at 3dp, and is strictly < 1.")
