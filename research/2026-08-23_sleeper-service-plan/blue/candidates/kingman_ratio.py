#!/usr/bin/env python3
# R2-1: distinguish shape-robustness from magnitude-dependence in the cadence queue.
# Kingman's G/G/1 approximation for mean wait in queue:
#   Wq ~= ((Ca^2 + Cs^2) / 2) * (rho / (1 - rho)) * (1/mu)
# M/M/1 is the special case Ca^2 = Cs^2 = 1, giving factor 1 (exact M/M/1 Wq).
# Calendar-timed (deterministic) arrivals => Ca^2 -> 0; exponential-scale service => Cs^2 = 1,
# giving factor (0+1)/2 = 0.5. Claim under test: the ratio is EXACTLY 1/2 at EVERY rho
# (magnitude scales by a constant), while the rho/(1-rho) shape and its rho=1 divergence
# are preserved for any nonzero (Ca^2 + Cs^2) (shape is distribution-free).
from fractions import Fraction as F

def kingman_factor(ca2, cs2):
    return F(ca2 + cs2, 2)

def wq(rho, mu, ca2, cs2):
    # returns Fraction multiples of the base rho/(1-rho)/mu times the Kingman factor
    return kingman_factor(ca2, cs2) * F(rho) / (1 - F(rho)) / F(mu)

mu = 3  # promotions/week; cancels in the ratio, any positive value works
rows = []
ratios = set()
for rho in [F(1,2), F(7,10), F(9,10), F(93,100), F(99,100), F(999,1000)]:
    mm1 = wq(rho, mu, 1, 1)          # M/M/1: Ca^2=Cs^2=1
    cal = wq(rho, mu, 0, 1)          # calendar arrivals: Ca^2=0, Cs^2=1
    ratio = cal / mm1
    ratios.add(ratio)
    rows.append((float(rho), float(mm1), float(cal), ratio))
    print(f"rho={float(rho):.3f}  Wq(M/M/1)={float(mm1):.4f}  Wq(cal,Ca2=0)={float(cal):.4f}  ratio={ratio}")

assert ratios == {F(1,2)}, f"ratio not constant 1/2: {ratios}"
print("RESULT: ratio is EXACTLY 1/2 at every rho tested (constant magnitude scaling).")

# Shape check: rho/(1-rho) diverges at rho->1 regardless of the (constant) Kingman factor.
for ca2, cs2 in [(0,1),(1,1),(2,3),(1,4)]:
    near = wq(F(999,1000), mu, ca2, cs2)
    print(f"Ca2={ca2} Cs2={cs2}: Wq(rho=0.999)={float(near):.2f}  -> diverges as rho->1 (shape preserved)")
print("RESULT: rho=1 divergence and rho/(1-rho) shape hold for every nonzero (Ca2+Cs2).")
