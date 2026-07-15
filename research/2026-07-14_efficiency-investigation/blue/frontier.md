# blue frontier — Efficiency and termination levers for the frank-exchange-of-views debate engine

Run 4 (efficiency investigation), 2026-07-14. Evidence base pinned per `inputs/PINNED.md`:
run-3 retrospective @ `bfa8a3b` (report §3 docket, cost.md, friction.md), engine + backlog @
`5396952`. Winnow list honored: PR #14–#18 items are shipped, not re-recommendable.

One hypothesis per docketed lever. Each states what would be TRUE if ratifying the lever is
right, the prediction it makes against the pinned corpus, and the disconfirming test that
would justify rejection. Doctrine constraint binds all five: cheapen redundancy and
mechanics, never judgment or the adversary; the spot-check floor never reaches zero.

## H1 — Severity-floor termination (ratify iff late rounds bought only trivia)

If severity-floor termination is right, run 3's record will show a round boundary after which
EVERY open gap was ≤ medium severity with low fix cost, and no later round surfaced a gap
above that floor — so routing the residual board to the judge at the floor would have saved
the late-round spend (cost.md: rounds 3–5 closed ~15 mostly-trivial gaps for ~the same $60
that rounds 1–2 spent closing 31) at zero verdict cost.
**Disconfirming test:** run 3's own late rounds minted R4-1 (lineage-blind docket, graded
High likelihood × High impact) and R5-5 (unenforced supersedes, HIGH) in rounds 4–5 — AFTER
the point the backlog claims the floor would have fired ("would have ended run 3 at round 3").
If those discoveries were reachable only by the extra full rounds, the floor as specified
terminates exactly the rounds that produced the corpus's highest-graded engine findings, and
the lever must be rejected or re-scoped (e.g. floor arms only when no NEW gap was minted that
round, not merely when open gaps are all low-severity).

## H2 — Risk-mass-proportional spend (ratify iff mass tracks value and the floor holds)

If risk-mass-proportional spend is right, sum(likelihood × impact) over open gaps computed at
each run-3 merge will correlate with the next round's realized gap-closure value (high mass
rounds 1–2 → high value; low mass rounds 3–5 → trivia), and the cost driver it throttles
(lens count × corpus size; $9.22–$11.05/round for 5 lenses, rising with corpus size while the
gap board shrank 20 → 6) is redundancy/mechanics, not judgment — so scoping lens passes to
mass cuts the dominant recurring line item without touching red-merge or the judge.
**Disconfirming tests:** (a) grades are red's own estimates — if run-3 grade corrections
(R2-1 count inflation, R3-7 mechanism narrowing, R5-1 discarded enumeration) moved computed
mass materially, the throttle input is too noisy to drive spend; (b) if any high-severity
late discovery (R4-1, R5-5) came from a lens pass a mass-scoped round would have cut (vs.
the always-on spot-check floor), the lever cheapens the adversary — doctrine violation, reject
or raise the floor.

## H3 — Blue grade-dispute channel; best-of-N grading only on evidence of surviving bias

If the structured grade-dispute channel is right, run 3's transcript will show blue already
disputing red's likelihood/impact grades in prose with real corrections landing (item 15's
likelihood re-grade cycle R1-13→R2-1→R3-7; impact re-grades R2-9, R5-2) but with no
machine-readable path — so disputes rode the general gap loop, invisible to the docket
detector, and a red re-rejection could persist without judge escalation (run 3: judge
dispatched ZERO times in 5 rounds). A `grade_disputes` envelope field + auto-docket on
re-rejection routes exactly this to the existing judge at near-zero mechanism cost.
If best-of-N grading is ALSO needed, the record will show lone-voice grade bias that the
adversarial loop did NOT catch — systematic over/under-grading surviving to the final report.
**Disconfirming test for best-of-N:** run 3's grade errors were all caught and corrected
within the adversarial loop itself (blue caught red's miscounts; red caught blue's stale
cells) — if no surviving-bias instance exists in the pinned corpus, best-of-N is scope creep:
adversarial-first suffices, defer the panel until runs 4–5 produce per-gap records showing
bias that survived (the backlog's own stated condition).

## H4 — Sharded findings (open ledger vs closed archive) + collator stage (ratify iff the full-re-read principle covers red-vs-blue, not red-vs-own-archive)

If sharding + collator are right, the merge seat's measured cost driver is TURNS × CONTEXT
(red-merge-r1: ~100–150K of material re-read every tool call, 2.7M+ cache reads; friction #15:
the 54KB findings.md forced three windowed reads + greps under the 25k Read cap), and the
full-re-read MUST in `red-auditor.md` protects the red-reads-BLUE adversarial check — not red
re-reading its own already-closed cases — so an open-items ledger + closed archive (merge
reads open + this round; archive remains readable on demand) plus a bulk-tier collator that
concatenates/normalizes lens passes into one digest cuts merge turns without cheapening any
judgment. Precedent predicted to hold: the citation ledger uses the identical
closed-items-don't-reopen pattern and held all prior confidences through round 4 with zero
observed regressions.
**Disconfirming tests:** (a) any run-3 catch that required red to re-read a CLOSED case of
its own (R5-1 was caught by reading red/findings.md status lines verbatim — determine whether
those lines would sit in the open ledger or the closed archive at that moment); (b) the
collator is a new seat between lenses and merge — if lens-pass nuance lost in normalization
(compound grades, lens-scoped labels, quoted anchors) would have changed a run-3 merge
outcome, the digest cheapens the adversary's input, not mechanics; (c) sanctioned write path:
sharding doubles red's living artifacts and every red write currently pays the
filename-keyed write-block copy detour (friction #4, #8, #10) — if the block hits the new
ledger file too, the lever's cost cell was undercounted.

## H5 — Round-scoped audit (held for cause: ratifiable only if scoping is propagation-aware, else reject)

If round-scoped audit (rounds 2+: changed sections + contested gaps + spot checks) is right,
every regression red actually caught by full re-read in run 3 would still have been caught
under the scoping rule — i.e. each one sat in a changed section, a contested lineage, or
plausible spot-check range. The corpus predicts the OPPOSITE for one class: blue's dominant
failure mode was incomplete propagation (5 chains in 5 rounds — R1-5→R2-4→R3-4/R3-9,
R2-8→R3-5→R4-3 …), and an unpropagated site is by definition in a section blue did NOT
change this round — a changed-sections audit is structurally blind to exactly the regression
class that cost run 3 the most rounds. R4-3 is the type specimen: the defective sentence was
one R3-5 left UNEDITED; no changelog entry pointed at it.
**Ratification condition (what would make it safe):** the scoping rule is propagation-aware —
for every correction accepted this round, ALL sites stating the corrected claim enter the
round's audit surface (a grep-cheap mechanical expansion, not judgment), the spot-check floor
stays nonzero and includes unchanged sections, and PR #15's blue-side propagation clause gets
at least one live run of evidence before red's backstop is narrowed. If run 4's own live
evidence shows the shipped propagation clause failing (winnow-list audit trigger), reject
round-scoping outright for run 5: it removes the only check that catches the engine's
measured dominant regression class.

## Cross-cutting priors (kept honest during search)

- Termination levers (H1, H2) attack round COUNT; structure levers (H4) attack per-round
  turns × context; H5 attacks per-round read volume. cost.md says cache traffic is 99% of
  tokens and judgment-seat rates (5×/12.5×) drive the premium — so H4's merge-seat lever
  targets the priciest seats ($10.60–$13.56/round red-merge), while H1/H2 target whole-round
  elimination ($25–30/round). Expect the ratify/reject calculus to differ accordingly.
- At least one search in five hunts disconfirming evidence; each hypothesis above names its
  disconfirming test up front so the searches are pre-committed.
