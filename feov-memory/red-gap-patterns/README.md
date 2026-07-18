# red's gap-pattern corpus — the compiled source, not the scratchpad

These are the patterns red has learned across runs 2-5: what kinds of defect
recur, where they hide, and what catches them. They are TRACKED, deliberately.

**Why tracked.** Memory here is not reading material, it is a source that gets
COMPILED INTO DUTY LINES that bind seats at the moment they act. E0.5 measured
the difference and it is stark: run 4's "read red's accumulated gap patterns"
clause was unsatisfiable at four blue seats, and run 5 was worse — lanes
verifiably READ the file and committed both warned patterns anyway. Only red's
duty-embedded patterns caught anything. Content that binds behaviour must be
reviewable and versioned, because an unreviewed input that binds behaviour is
exactly the poisoning surface the constitutional reform names.

**Why here and not in `.claude/agent-memory/`.** That directory is the harness's
agent-memory path: machine-local, resolved through `CLAUDE_PROJECT_DIR`, and
gitignored. It nearly cost us the whole corpus once — the 2026-07-18 cwd move
off AgentOrange would have started red amnesiac if the files had not been carried
across by hand. A corpus that survives only on one developer's disk is not a
corpus. Clone the repo, get the memory.

**The two tiers.**

- `feov-memory/red-gap-patterns/` (here) — PROMOTED patterns. Reviewed, tracked,
  and what run setup mirrors into `inputs/red-gap-patterns.md`.
- `.claude/agent-memory/frank-exchange-of-views-red-auditor/` — raw accrual. Red
  writes here during a run. Ignored by git. Promotion into this directory is a
  deliberate review step, not an automatic sync: it is the moment a human decides
  a pattern is worth binding future seats to.

## How these are delivered (memory-as-duty, shipped)

Each pattern carries `metadata.classes` in its frontmatter, naming classes from
`feov-memory/class-registry-seed.md`. Setup builds
`inputs/gap-patterns-by-class.json`; the engine hands a repairing seat ONLY the
patterns whose class matches the gap in front of it, and blue's manifest row
records which patterns it checked and what checking them showed.

**Why a join and not a search.** A seat handed some patterns by relevance-search
leaves nobody able to say whether they were the right ones — there is nothing for
a detector to key on, so the scorecard cannot tell a discharged duty from a
skipped one. A class join is deterministic and auditable, and it needs no index.

**Why not stage the whole corpus.** That is what we used to do, and E0.5 measured
it: run 4's read-the-patterns clause was unsatisfiable at four blue seats, and
run 5's lanes verifiably read the staged file and committed both warned patterns
anyway. Reading is not binding. Fifty patterns at seat start is a salience
problem that no amount of instruction fixes.

**An unclassified pattern is not delivered.** Classification happens at
promotion, which is the moment a human is already looking at the pattern and
deciding whether future seats should be bound by it. A pattern with no class is a
promotion that skipped its review step, not a pattern that applies everywhere.

**Still instance-shaped.** These entries describe specific incidents, and a class
is a kind. The classes make delivery work; making the ENTRIES themselves
class-shaped — so blue cannot discharge the duty by matching the listed instance
and calling it done — is the remaining half of the argument in
`ideas/gap-classes-proposal.md`.
