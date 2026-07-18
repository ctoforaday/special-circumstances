---
name: pattern-self-defeating-mitigation
description: A control added to close a prior-round gap introduces its own failure mode — audit every new mitigation as a fresh attack surface, not a closed ticket
metadata:
  classes: [unverified-composition, partial-control-coverage, cross-section-contradiction]
  type: feedback
---

When blue adds a mitigation in round N to close a round N-1 gap, do NOT treat the gap as closed on
acceptance — re-audit the mitigation itself as new surface. Recurring sub-patterns observed:

- **Control collides with the system's own write loop.** A clone-injection defense that gates
  activation on a content-fingerprint marker breaks because the nightly `/dream` pass *mutates the
  store*, invalidating the fingerprint every run — forcing either self-ratification by the unattended
  pass (defeats the gate) or daily manual re-ratify (unworkable). Ask: does the new control's
  invariant survive the system's own routine writes?
- **Control depends on the very diligence the report elsewhere discredits.** §2.4 demoted human
  diff-review to "forensic" because it decays to LGTM; then the clone defense makes human ratification
  the *only* preventive control. A mitigation cannot lean on an assumption the same report argues is
  unreliable.
- **Control has an escape hatch that reopens the common case.** "auto-ratify repos under a configured
  trusted root" voids the clone defense for the solo-dev workflow (clone everything under ~/Projects).
- **Control closes the durable path but not the in-pass path.** Ephemeral consolidator memory closes
  *durable* self-poisoning but the agent still *reads* the poisonable store each pass → in-pass
  steering residual. "Sits outside the trust surface" is overstated when it still ingests the guarded
  surface.
- **Verification channel is the same as the injection channel.** A risk-accept names "independent
  re-verification" as the mitigant for content-poisoning, but the actual leaf-node protocol is
  "follow the citation to the source" — i.e. re-fetch the *same* URL, not a second independently-found
  one. A single poisoned/fabricated source that is internally consistent gets checked by re-reading
  itself; that isn't independence. Grep the protocol text for the word the mitigation leans on
  ("independent") before crediting it — if it doesn't appear, the mitigation is asserted, not built.
- **A two-party dispute/correction channel audits only its contested branch.** A grade-dispute
  mechanism dockets a dispute to the judge only after reject → re-dispute; a dispute the counterparty
  ACCEPTS silently rewrites the permanent record with no log, docket, or spot-check — and acceptance
  is the cheap/colluding path (both parties have budget stakes under actuation). Companion holes: the
  loop's PASS/terminal break precedes the dispute filter (ending the run moots pending disputes —
  the channel fails the final-round case it exists for), and dispute traffic is uncapped (a cost
  lever pointed at the responder). Ask of any adjudication channel: what happens on the AGREE path,
  at termination, and at volume? An incentive analysis that grades only one party's incentive is
  half an analysis.
- **Guard scheduled after the commitment it gates.** A write-guard preflight repaired twice
  (simulator→live seat, then live-seat→same seat class) ends up scheduled as "the first sharded
  run's red-merge writes both shards as its opening act" — i.e. it can only FAIL after the PR
  shipping the sharded design (renamed prompts, retired monolith) is already merged. A preflight
  that fires past the point of no return is a smoke alarm wired to the ashes; also check the fix
  didn't contradict a sibling clause (here: "skeleton-born" names a different creator seat).
  Ask of every preflight/gate: at the moment this check can first fail, is the decision it
  guards still reversible? (efficiency-investigation round 4, L6-F3 — vector was red's own
  R3-11 required-fix example, repeated by the lead.)
- **Two controls added in different rounds silently collide.** A cost-saving cache ("verified
  citations don't un-verify unless the citing prose changed") and a drift-mitigation ("record
  access-date deltas because sources change") get shipped separately, each reads fine alone, but the
  cache's skip-condition is keyed to internal edit history while the drift risk is external and
  time-based — so the cache actively prevents the re-check the drift fix's own rationale ("drift is
  usually caught by re-verification") depends on. Neither section mentions the other. Ask, for every
  pair of controls addressing adjacent risk classes: does control A's skip/short-circuit condition
  starve control B of the trigger it needs?

**Why:** blue's Round-1 additions were directionally correct and accepted, but each carried an
un-graded second-order failure — accepting the direction is not accepting the implementation.

**How to apply:** for every ACCEPT in blue's response section, write the adversarial one-liner "how
does this new mechanism fail / get bypassed / collide with existing machinery" before crediting the
gap as closed. Grade the residual separately.
