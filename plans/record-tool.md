# The record tool: physical guarantees instead of legislated discipline

Proposed (user, 2026-07-18) after the E0.5 convergence: four independent audits
acquitted the seats and indicted the record (attestations honest but phrased
unauditably; merge lossless but retention lossy; boards overwritten in place;
note-class observations with no fate). The response: the run record becomes an
EVENT LOG owned by a tool, and everything downstream — markdown boards, telemetry,
scorecards, audits — becomes a PROJECTION. Absorbs/reshapes W2c(record parts),
W2d, W2f, W2h. Route: this plan → plan-audit gate → implementation PRs.

## The five goals (the proposal, verbatim intent)

1. AUTOMATIC METRICS: every scorecard number is a projection of events — never a
   seat self-report, never an assembly-minted aggregate (the E0.5f defect class
   dies at the root: assembly copies projections, computes nothing).
2. LOSSLESS BY CONSTRUCTION: append-only event log; the tool has no mutation or
   deletion operation. In-place overwrite (which destroyed efficiency rounds 1-3
   and made 19% of sleeper round-4 grade decisions unauditable) becomes
   physically impossible. Any historical board state = replay to seq N.
3. STRUCTURED TAGGING: class (validated against the seed registry, tie-break
   hints printed on ambiguity), grades (enum-checked), lineage (supersedes must
   reference an EXISTING id — the tool refuses dangling edges), found_by at
   lens-finding granularity, contributing grades beside minted grades.
4. PROBLEM-CLASS PREVENTION, mapped to measured findings: anchor-required
   closures = required --seat/--tool/--target args (E0.5a); note-fate tracking =
   lens observations demand a merge disposition (E0.5b); id collisions = ids
   minted by the tool (cross-corpus-id-collision class); count self-reports
   (ledger_closure_lines/archive_blocks) deleted — the tool counts.
5. CONSTITUTIONAL DELETION: procedure leaves the constitutions; judgment stays.

## Shape

`record.mjs` (FEOV scripts; node is FEOV's required dep — same test harness as
the other scripts). Storage: `<runDir>/records/events.jsonl`, append-only, one
event per line: {seq, seat, type, payload}. Seats invoke subcommands:

  record.mjs mint    --seat red-merge-r2 --class enumeration-non-exhaustive \
                     --location "…" --problem "…" --fix "…" --check "…" \
                     --severity mh --likelihood m --impact mh \
                     --found-by L5-F3,L6-F2 [--supersedes R1-16]
  record.mjs close   --id R1-16 --class-of-closure closed_with_regression \
                     --anchor-seat L1 --anchor-tool "git show" --anchor-target "7bc501e:…" \
                     --successor R2-4 [--carried-from r2]   # carried-as-fresh becomes impossible to phrase
  record.mjs observe --seat L4 --kind note --text "…"       # merge later: record.mjs dispose --obs 17 --as declined --reason "…"
  record.mjs regrade --id R2-5 --likelihood l --basis "…"   # same-id regrades become possible AND recorded
  record.mjs friction / opinion / petition / manifest-row / spot-check / revision …

  record.mjs render ledger|archive|telemetry|scorecards|judicial-record
    → ledger.md, archive.md, board-telemetry.jsonl, the dashboard model, and the
      report's aggregate sections become GENERATED VIEWS. Seats stop hand-writing
      boards entirely; humans keep reading markdown; audits read events.

Validation at append time: enums, registry classes, existing-id references,
required anchors. The log is git-tracked; capture's audit collapses to "log
parses; views match a fresh render" — most current heuristic audits retire.

## The constitutional deletion list (the payoff)

DELETED from red: ledger/archive maintenance mechanics (sharding format, closure
index shape, NEVER-edit-an-archive-block, count self-reports, near-match index
discipline as prose), the telemetry line spec + formulas, the attestation-format
invariant prose (anchors are now arguments), spot-check recording mechanics.
DELETED from blue: CHANGELOG round-entry mechanics (revision events), manifest
envelope plumbing (manifest-row events). DELETED from the bench: judicial-record
assembly mechanics (rendered view), the recompute-or-cite gate (nothing left to
recompute — assembly copies projections). RETAINED everywhere: telos, win/loss,
craft duties, grading judgment, what to write in the prose fields. The
constitutions state WHO records WHAT and WHY; the tool owns HOW.

## What stays honest

- Prose payloads (problem statements, rationales, closings) remain prose — the
  tool structures the envelope, never the argument.
- The tool cannot verify a claim, only its shape: vacuity's auditor remains
  post-hoc behavior audits (tool-call index), now trivially joinable to events.
- Migration: v1 runs the tool ALONGSIDE generated views double-checked against
  the old prompts' outputs for one run; prompts shrink in v2 after parity holds.
- The event log is the record layer W2f promised, promoted from convention
  (agents append JSONL politely) to enforcement (the only write path is the tool).

## Sequencing

R1: record.mjs core (mint/close/observe/dispose/regrade/friction + render
    ledger/archive/telemetry) + simulator tests. Rides the W2f slot.
R2: engine prompts switch seats to tool invocations; constitutional deletions
    land the same PR (the prompt shrink IS the review surface).
R3: scorecards/judicial-record/dashboard-model renders (absorbs W2h + W2c record
    parts); capture audit collapse.
R4: class registry live (W2d) with recalibrated within-run escalator semantics.
