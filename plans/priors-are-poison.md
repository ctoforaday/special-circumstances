# Priors are poison — a chair reads its OWN in-run scorecard, not a cross-run seed

> STATUS 2026-09-02: shipped — historical record. All three fix items landed (commit 010f6697; debate.js scorecardClause is now the in-run instruction; the seat-facing read shipped as `show scorecard` rather than `show --view scorecard`; the `scorecards` arg feeds operator analytics only).

Ruling 2026-07-19 (smoke loop). **debate.js chair-prompt change + one tool addition. Queue
with the after-run batch, near the top** — editing debate.js mid-run breaks replay.

## The finding

Each chair's prompt is seeded (`scorecardClause`) with the PRIOR run's scorecard headlines
— its "memory". On the smoke dashboard this showed a populated bench and a
`repair_regression_ratio 0.67` before anything had been repaired. That surfaced the deeper
problem: seeding a chair with cross-run scores is not neutral, it is likely HARMFUL.

- **Goodhart.** The scorecard is the INSTRUMENT that measures the chair. Show a chair its
  own `anchored_closures_pct 0` and its cheapest move is to add the anchor fields, verified
  or not — it performs the metric instead of doing the thing. You corrupt the behaviour AND
  the instrument's validity at once. This is self-report one level up — the exact
  anti-pattern the 2026-07-19 tool migration removed from the merge envelope.
- **Cross-topic confound.** A metric like `citation_yield` measures the TASK, not the chair:
  a local-file/reasoning topic never exercises web fetch and scores clean; an
  internet-research topic leans on it and scores badly. A prior citation score shown on a
  different topic misdirects effort toward citations the current question does not need.
- **Cross-model overreach.** We seed a haiku chair with an opus run's numbers — priming a
  competence profile it cannot execute.
- **Salience.** "You failed at X" makes X more available (the gap-patterns were read and the
  warned patterns committed anyway — priming the failure, not preventing it).

No evidence of benefit; three concrete mechanisms for harm; the floor case is wasted tokens.

## The distinction that matters

- **In-run behaviour on THIS question = legitimate.** A chair should see how it is actually
  doing on the problem in front of it; if red is gating blue heavily, blue should recognise
  it is failing and adapt.
- **Cross-run priors from OTHER questions = poison.** Remove them.

And the rule for HOW to give in-run feedback: the good signal is the **adversary's actual
verdict** (red's real gates — Goodhart-RESISTANT, because red re-verifies at the leaf and
the chair cannot game it by performing), NOT a computed performance metric shown to the
chair (gameable). Metrics are for the OPERATOR, who cannot game them.

## The fix (small)

Not a big prompt excision — a mechanism swap:

1. **Replace** `scorecardClause`'s seed-injection with an INSTRUCTION: before reading the
   open docket, run `feov-record <role> show --view scorecard` to read your performance THIS
   run. A fresh seat sees a blank slate; the scorecard fills from this run's record as the
   run proceeds. Active pull of ground truth, not a pushed cross-run seed (see
   tool-is-the-contract.md §IX).
2. **Add** the tool capability: `show --view scorecard` computes the CALLING role's
   scorecard live from the record (the board + events), for this run only. Blank until there
   is in-run history to compute.
3. **The `scorecards` workflow arg stops feeding the chairs.** It may still feed post-run
   OPERATOR analytics (across runs, where a stable trait could actually emerge and no one can
   game it) — but never the chair prompts.

## Goodhart caveat (keep it honest)

An in-run scorecard read is safest for metrics the tool ENFORCES (ungameable — you cannot
fake an anchor the tool requires) and for descriptive ones. The residual risk is a chair
chasing an un-enforced in-run number; enforcement keeps shrinking that surface. Prefer
surfacing the raw adversarial record ("red has gated N of your claims this run", read from
the board) over a computed score wherever the behaviour is not yet enforced.
